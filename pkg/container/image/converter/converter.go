package converter

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/darxkies/k8s-tew/pkg/container/image/manifest"
	"github.com/darxkies/k8s-tew/pkg/container/image/storage"
	digest "github.com/opencontainers/go-digest"
	ociv1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/smallnest/goreq"
)

const DOCKER_DISTRIUBUTION_MANIFEST_LIST = "application/vnd.docker.distribution.manifest.list.v2+json"
const DOCKER_DISTRIUBUTION_MANIFEST = "application/vnd.docker.distribution.manifest.v2+json"

type imageConverter struct {
	domain    string
	imageName string
	tag       string
	token     string
	debug     bool
	layers    layers
	storage   storage.Storage
}

func (converter *imageConverter) getManifestAddress(tag string) string {
	return fmt.Sprintf("https://%s/v2/%s/manifests/%s", converter.domain, converter.imageName, tag)
}

func (converter *imageConverter) getBlobAddress(blob string) string {
	return fmt.Sprintf("https://%s/v2/%s/blobs/%s", converter.domain, converter.imageName, blob)
}

func (converter *imageConverter) stripQuotes(value string) string {
	if len(value) > 0 && value[0] == '"' {
		value = value[1:]
	}

	if len(value) > 0 && value[len(value)-1] == '"' {
		value = value[:len(value)-1]
	}

	return value
}

func (converter *imageConverter) getToken() error {
	httpClient := &http.Client{}

	address := converter.getManifestAddress(converter.tag)

	log.WithFields(log.Fields{"address": address, "image": converter.String()}).Debug("Image converter token")

	response, body, error := goreq.New().SetClient(httpClient).
		Get(address).
		SetDebug(converter.debug).
		End()

	if error != nil && len(error) > 0 {
		return error[0]
	}

	if response.StatusCode == 200 {
		return nil
	}

	if response.StatusCode != 401 {
		return fmt.Errorf("No 401 received")
	}

	authenticate := response.Header.Get("Www-Authenticate")

	tokens := strings.Split(authenticate, " ")

	if len(tokens) != 2 {
		return fmt.Errorf("Wrong www-authenticate format '%s'", authenticate)
	}

	if tokens[0] != "Bearer" {
		return fmt.Errorf("No Bearer found in '%s'", authenticate)
	}

	tokens = strings.Split(tokens[1], ",")

	realm := ""
	service := ""
	scope := ""

	for _, token := range tokens {
		keyValue := strings.Split(token, "=")

		if len(keyValue) != 2 {
			return fmt.Errorf("Malformated key/value '%s' in '%s'", token, authenticate)
		}

		key := keyValue[0]
		value := keyValue[1]

		value = converter.stripQuotes(value)

		if key == "realm" {
			realm = value
		}

		if key == "service" {
			service = value
		}

		if key == "scope" {
			scope = value
		}
	}

	if len(realm) == 0 {
		return fmt.Errorf("No realm in '%s'", authenticate)
	}

	if len(service) == 0 {
		return fmt.Errorf("No service in '%s'", authenticate)
	}

	if len(scope) == 0 {
		return fmt.Errorf("No scope in '%s'", authenticate)
	}

	address = fmt.Sprintf("%s?service=%s&scope=%s", realm, service, scope)

	log.WithFields(log.Fields{"address": address, "image": converter.String()}).Debug("Image converter token")

	response, body, error = goreq.New().SetClient(httpClient).
		SetBasicAuth("", "").
		Get(address).
		SetDebug(converter.debug).
		End()

	if error != nil && len(error) > 0 {
		return error[0]
	}

	if response.StatusCode != 200 {
		return fmt.Errorf("No 200 received  for token request")
	}

	var tokenData map[string]*json.RawMessage

	if _error := json.Unmarshal([]byte(body), &tokenData); _error != nil {
		return _error
	}

	token, ok := tokenData["token"]

	if !ok {
		return fmt.Errorf("No token found in the response")
	}

	converter.token = converter.stripQuotes(string(*token))

	log.WithFields(log.Fields{"token": string(converter.token), "image": converter.String()}).Debug("Image converter token")

	return nil
}

func (converter *imageConverter) String() string {
	return fmt.Sprintf("%s/%s:%s", converter.domain, converter.imageName, converter.tag)
}

func (converter *imageConverter) downloadDefinition(tag string) (string, string, error) {
	httpClient := &http.Client{}

	address := converter.getManifestAddress(tag)

	log.WithFields(log.Fields{"address": address, "image": converter.String()}).Debug("Image converter download definition")

	request := goreq.New().SetClient(httpClient).
		Get(address).
		SetDebug(converter.debug).
		SetHeader("Accept", strings.Join([]string{ociv1.MediaTypeImageIndex, ociv1.MediaTypeImageManifest, DOCKER_DISTRIUBUTION_MANIFEST_LIST, DOCKER_DISTRIUBUTION_MANIFEST}, ", "))

	if len(converter.token) > 0 {
		request = request.SetHeader("Authorization", "Bearer "+converter.token)
	}

	response, body, error := request.End()

	if error != nil {
		return "", "", fmt.Errorf("Download from '%s' failed: %w", address, error[0])
	}

	if response.StatusCode != 200 {
		log.WithFields(log.Fields{"address": address, "image": converter.String(), "body": body}).Error("Image converter download definition failed")

		return "", "", fmt.Errorf("Could not download manifest for '%s' (%s)", converter.String(), body)
	}

	digest := response.Header.Get("docker-content-digest")

	return body, digest, nil
}

func (converter *imageConverter) getBlob(layer *layer, skipCheck bool, isGzip bool) error {
	log.WithFields(log.Fields{"filename": layer.Filename, "image": converter.String()}).Debug("Image converter downloading blob")

	if layer.BlobDigest == manifest.EmptyLayer {
		return nil
	}

	httpClient := &http.Client{}

	address := converter.getBlobAddress(layer.BlobDigest)

	request := goreq.New().SetClient(httpClient).
		Get(address)

	if len(converter.token) > 0 {
		request = request.SetHeader("Authorization", "Bearer "+converter.token)
	}

	response, body, error := request.EndBytes()

	if error != nil {
		return fmt.Errorf("Blob download from '%s' failed: %w", address, error[0])
	}

	if response.StatusCode != 200 {
		return fmt.Errorf("Could not download blob '%s' for '%s', (%s)", layer.BlobDigest, converter.String(), body)
	}

	if error := converter.storage.WriteFile(layer.Filename, body); error != nil {
		return error
	}

	if skipCheck {
		return nil
	}

	// Check digest of downloaded blob
	digest := fmt.Sprintf("sha256:%x", sha256.Sum256(body))

	if layer.BlobDigest != digest {
		return fmt.Errorf("Layer %s of %s could not be downloaded properly", layer.BlobDigest, converter.String())
	}

	// Get blob size
	layer.Size = int64(len(body))

	// Get digest of blob tar
	{
		buffer := bytes.NewReader(body)
		hash := sha256.New()

		if isGzip {
			zw, error := gzip.NewReader(buffer)
			if error != nil {
				return fmt.Errorf("GZIP open for '%s' failed: %w", digest, error)
			}

			if _, error := io.Copy(hash, zw); error != nil {
				return fmt.Errorf("GZIP copy for '%s' failed: %w", digest, error)
			}
		} else {
			if _, error := io.Copy(hash, buffer); error != nil {
				return fmt.Errorf("TAR copy for '%s' failed: %w", digest, error)
			}
		}

		layer.TarDigest = fmt.Sprintf("sha256:%x", hash.Sum(nil))
	}

	return nil
}

func (converter *imageConverter) process() error {
	log.WithFields(log.Fields{"image": converter.String()}).Debug("Image converter")

	if error := converter.getToken(); error != nil {
		return error
	}

	body, dockerDigest, error := converter.downloadDefinition(converter.tag)

	if error != nil {
		return error
	}

	ociIndex := ociv1.Index{}

	if _error := json.Unmarshal([]byte(body), &ociIndex); _error == nil {
		log.WithFields(log.Fields{"schema-version": ociIndex.SchemaVersion, "media-type": ociIndex.MediaType, "digest": dockerDigest}).Debug("Image converter index metadata")

		if ociIndex.SchemaVersion == 2 && (ociIndex.MediaType == ociv1.MediaTypeImageIndex || ociIndex.MediaType == DOCKER_DISTRIUBUTION_MANIFEST_LIST) {
			return converter.saveOCI(&ociIndex)
		}
	}

	ociManifest := ociv1.Manifest{}

	if _error := json.Unmarshal([]byte(body), &ociManifest); _error == nil {
		log.WithFields(log.Fields{"schema-version": ociManifest.SchemaVersion, "media-type": ociManifest.MediaType, "digest": dockerDigest}).Debug("Image converter manifest metadata")

		if ociManifest.SchemaVersion == 2 && ociManifest.MediaType == DOCKER_DISTRIUBUTION_MANIFEST {
			return converter.saveOCIManifest(&ociManifest)
		}
	}

	return fmt.Errorf("unknown image format in %s", converter.String())
}

func (converter *imageConverter) writeOCILayout() error {
	ociLayout := ociv1.ImageLayout{Version: ociv1.ImageLayoutVersion}

	data, error := json.Marshal(&ociLayout)
	if error != nil {
		return error
	}

	filename := "oci-layout"

	return converter.storage.WriteFile(filename, data)
}

func (converter *imageConverter) writeOCIIndex(digest digest.Digest, size int) error {
	ociIndex := ociv1.Index{}
	ociIndex.SchemaVersion = 2
	ociIndex.MediaType = ociv1.MediaTypeImageIndex

	ociManifest := ociv1.Descriptor{}
	ociManifest.MediaType = ociv1.MediaTypeImageManifest
	ociManifest.Digest = digest
	ociManifest.Size = int64(size)
	ociManifest.Annotations = map[string]string{
		ociv1.AnnotationRefName: converter.tag,
	}

	ociIndex.Manifests = append(ociIndex.Manifests, ociManifest)

	data, error := json.Marshal(&ociIndex)
	if error != nil {
		return error
	}

	filename := "index.json"

	return converter.storage.WriteFile(filename, data)
}

// PullImage downloads an image from a repository and converts it to an OCI archive
func PullImage(imageName, outputFilename string, debug bool) error {
	storage, error := storage.NewTarStorage(outputFilename)
	if error != nil {
		return error
	}

	defer storage.Close()

	tokens := strings.Split(imageName, "/")

	if len(tokens) != 2 && len(tokens) != 3 {
		return fmt.Errorf("'%s' is not a valid image name", imageName)
	}

	registryClient := &imageConverter{storage: storage, tag: "latest", debug: debug}

	if tokens[0] == "docker.io" {
		registryClient.domain = "registry-1.docker.io"
	} else {
		registryClient.domain = tokens[0]
	}

	if len(tokens) == 2 {
		registryClient.imageName = tokens[1]
	} else {
		registryClient.imageName = tokens[1] + "/" + tokens[2]
	}

	tokens = strings.Split(registryClient.imageName, ":")

	if len(tokens) == 2 {
		registryClient.imageName = tokens[0]
		registryClient.tag = tokens[1]
	}

	if error := registryClient.process(); error != nil {
		_ = storage.Remove()

		return error
	}

	return nil
}

func (converter *imageConverter) saveOCI(ociIndex *ociv1.Index) error {
	var digest digest.Digest

	// Select manifest
	for _, manifest := range ociIndex.Manifests {
		log.WithFields(log.Fields{"media-type": manifest.MediaType, "digest": manifest.Digest, "architecture": manifest.Platform.Architecture, "os": manifest.Platform.OS, "image": converter.String()}).Debug("Image converter manifest")

		if manifest.Platform.Architecture == "amd64" && manifest.Platform.OS == "linux" {
			digest = manifest.Digest
		}
	}

	if len(digest) == 0 {
		return fmt.Errorf("Could not find a matching manifest in '%s'", converter.String())
	}

	// Download manifest
	body, _, error := converter.downloadDefinition(digest.String())
	if error != nil {
		return error
	}

	ociManifest := ociv1.Manifest{}

	if _error := json.Unmarshal([]byte(body), &ociManifest); _error != nil {
		return _error
	}

	return converter.saveOCIManifest(&ociManifest)
}

func (converter *imageConverter) saveOCIManifest(ociManifest *ociv1.Manifest) error {
	blobsDirectory := path.Join("blobs", "sha256")
	layer := &layer{}

	// Write Blobs
	for _, ociLayer := range ociManifest.Layers {
		layer.Filename = path.Join(blobsDirectory, ociLayer.Digest.Encoded())
		layer.BlobDigest = ociLayer.Digest.String()

		if error := converter.getBlob(layer, false, strings.Contains(ociLayer.MediaType, "gzip")); error != nil {
			return error
		}
	}

	// Write Config
	layer.Filename = path.Join(blobsDirectory, ociManifest.Config.Digest.Encoded())
	layer.BlobDigest = ociManifest.Config.Digest.String()

	if error := converter.getBlob(layer, true, false); error != nil {
		return error
	}

	// Patch Manifest
	ociManifest.MediaType = ociv1.MediaTypeImageManifest
	ociManifest.Config.MediaType = ociv1.MediaTypeImageConfig

	for i := range ociManifest.Layers {
		if strings.Contains(ociManifest.Layers[i].MediaType, "gzip") {
			ociManifest.Layers[i].MediaType = ociv1.MediaTypeImageLayerGzip
		} else {
			ociManifest.Layers[i].MediaType = ociv1.MediaTypeImageLayer
		}
	}

	// Write Manifest
	body, error := json.Marshal(ociManifest)
	if error != nil {
		return error
	}

	sum256 := sha256.Sum256(body)
	manifestDigest := digest.NewDigestFromBytes(digest.SHA256, sum256[:])

	filename := path.Join(blobsDirectory, manifestDigest.Encoded())

	if _error := converter.storage.WriteFile(filename, []byte(body)); _error != nil {
		return _error
	}

	// Write Index
	if _error := converter.writeOCIIndex(manifestDigest, len(body)); _error != nil {
		return _error
	}

	// Write OCI Layout
	converter.writeOCILayout()

	return nil
}
