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
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/darxkies/k8s-tew/pkg/container/image/manifest"
	"github.com/darxkies/k8s-tew/pkg/container/image/storage"
	digest "github.com/opencontainers/go-digest"
	ociv1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/smallnest/goreq"
)

type imageConverter struct {
	domain    string
	imageName string
	tag       string
	token     string
	debug     bool
	layers    layers
	manifest  *manifest.Manifest
	storage   storage.Storage
}

func (converter *imageConverter) getManifestAddress() string {
	return fmt.Sprintf("https://%s/v2/%s/manifests/%s", converter.domain, converter.imageName, converter.tag)
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

	address := converter.getManifestAddress()

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

func (converter *imageConverter) getManifest() error {
	httpClient := &http.Client{}

	address := converter.getManifestAddress()

	log.WithFields(log.Fields{"address": address, "image": converter.String()}).Debug("Image converter manifest")

	request := goreq.New().SetClient(httpClient).
		Get(address).
		SetDebug(converter.debug).
		SetHeader("Accept", "application/vnd.docker.distribution.manifest.v1+prettyjws")

	if len(converter.token) > 0 {
		request = request.SetHeader("Authorization", "Bearer "+converter.token)
	}

	response, body, error := request.End()

	if error != nil {
		return error[0]
	}

	if response.StatusCode != 200 {
		return fmt.Errorf("Could not download manifest for '%s' (%s)", converter.String(), body)
	}

	converter.manifest = &manifest.Manifest{}

	if _error := json.Unmarshal([]byte(body), converter.manifest); _error != nil {
		return _error
	}

	log.WithFields(log.Fields{"schema-version": converter.manifest.SchemaVersion, "name": converter.manifest.Name, "tag": converter.manifest.Tag, "architecture": converter.manifest.Architecture, "image": converter.String()}).Debug("Image converter manifest")

	return nil
}

func (converter *imageConverter) getBlob(layer *layer) error {
	log.WithFields(log.Fields{"filename": layer.Filename, "image": converter.String()}).Debug("Image converter downloading blob")

	if layer.BlobDigest == manifest.EmptyLayer {
		layer.EmptyLayer = true

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
		return error[0]
	}

	if response.StatusCode != 200 {
		return fmt.Errorf("Could not download blob '%s' for '%s', (%s)", layer.BlobDigest, converter.String(), body)
	}

	if error := converter.storage.WriteFile(layer.Filename, body); error != nil {
		return error
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

		zw, error := gzip.NewReader(buffer)
		if error != nil {
			return error
		}

		hash := sha256.New()
		if _, error := io.Copy(hash, zw); error != nil {
			return error
		}

		layer.TarDigest = fmt.Sprintf("sha256:%x", hash.Sum(nil))
	}

	return nil
}

func (converter *imageConverter) downloadLayers() error {
	blobsDirectory := path.Join("blobs", "sha256")

	for _, layer := range converter.layers {
		layer.Filename = path.Join(blobsDirectory, layer.BlobDigest[7:])

		if error := converter.getBlob(layer); error != nil {
			return error
		}
	}

	return nil
}

func (converter *imageConverter) process() error {
	log.WithFields(log.Fields{"image": converter.String()}).Debug("Image converter")

	if error := converter.getToken(); error != nil {
		return error
	}

	if error := converter.getManifest(); error != nil {
		return error
	}

	if len(converter.manifest.FileSystemLayers) != len(converter.manifest.History) {
		return fmt.Errorf("%s does not have the same number of layers and history entries", converter.String())
	}

	converter.layers = layers{}

	for i := len(converter.manifest.History) - 1; i >= 0; i-- {
		historyEntry := converter.manifest.History[i]

		data := &manifest.HistoryEntryData{}

		if _error := json.Unmarshal([]byte(historyEntry.V1Compatibility), &data); _error != nil {
			return _error
		}

		layer := &layer{}
		layer.BlobDigest = converter.manifest.FileSystemLayers[i].BlobSum
		layer.History = data
		layer.MediaType = "application/vnd.docker.image.rootfs.diff.tar.gzip"

		log.WithFields(log.Fields{"id": data.ID, "layer-parent": data.Parent, "docker-version": data.DockerVersion, "architecture": data.Architecture, "os": data.OS, "container": data.Container, "throwaway": data.Throwaway, "created": data.Created, "author": data.Author, "image": converter.String()}).Debug("Image converter layer")

		if data.Config != nil {
			data, _ := json.Marshal(data.Config)

			log.WithFields(log.Fields{"config": string(data), "image": converter.String()}).Debug("Image converter layer config")
		}

		if data.ContainerConfig != nil {
			data, _ := json.Marshal(data.ContainerConfig)

			log.WithFields(log.Fields{"config": string(data), "image": converter.String()}).Debug("Image converter layer container config")
		}

		converter.layers = append(converter.layers, layer)
	}

	if error := converter.downloadLayers(); error != nil {
		return error
	}

	hashImageConfig, size, error := converter.writeOCIImageConfig()
	if error != nil {
		return error
	}

	hashManifest, size, error := converter.writeOCIManifest(hashImageConfig, size)
	if error != nil {
		return error
	}

	if error := converter.writeOCIIndex(hashManifest, size); error != nil {
		return error
	}

	if error := converter.writeOCILayout(); error != nil {
		return error
	}

	log.WithFields(log.Fields{"manifest-blob": hashManifest[7:], "image": converter.String()}).Debug("Image converter")
	log.WithFields(log.Fields{"image-config-blob": hashImageConfig[7:], "image": converter.String()}).Debug("Image converter")

	return nil
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

func (converter *imageConverter) writeOCIIndex(hash string, size int) error {
	ociIndex := ociv1.Index{}
	ociIndex.SchemaVersion = 2

	digest, error := digest.Parse(hash)
	if error != nil {
		return error
	}

	ociManifest := ociv1.Descriptor{}
	ociManifest.Digest = digest
	ociManifest.MediaType = ociv1.MediaTypeImageManifest
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

func (converter *imageConverter) writeOCIImageConfig() (hash string, size int, error error) {
	ociImageConfig := &ociv1.Image{}

	timestamp, error := time.Parse(time.RFC3339, converter.layers[len(converter.layers)-1].History.Created)

	if error != nil {
		return
	}

	ociImageConfig.Created = &timestamp
	ociImageConfig.Architecture = converter.manifest.Architecture
	ociImageConfig.OS = converter.layers[len(converter.layers)-1].History.OS
	ociImageConfig.Config = ociv1.ImageConfig{}

	if len(converter.layers) > 0 {
		for _, layer := range converter.layers {
			if layer.History.Config == nil {
				continue
			}

			if layer.History.Config.Cmd != nil {
				ociImageConfig.Config.Cmd = layer.History.Config.Cmd
			}

			if layer.History.Config.Entrypoint != nil {
				ociImageConfig.Config.Entrypoint = layer.History.Config.Entrypoint
			}

			ociImageConfig.Config.WorkingDir = layer.History.Config.WorkingDir
			ociImageConfig.Config.Volumes = layer.History.Config.Volumes
			ociImageConfig.Config.ExposedPorts = layer.History.Config.ExposedPorts
			ociImageConfig.Config.Labels = layer.History.Config.Labels
			ociImageConfig.Config.Entrypoint = layer.History.Config.Entrypoint
			ociImageConfig.Config.Env = layer.History.Config.Env
			ociImageConfig.Config.User = layer.History.Config.User
		}
	}

	config, _ := json.Marshal(ociImageConfig.Config)

	log.WithFields(log.Fields{"config": string(config), "image": converter.String()}).Debug("Image converter config")

	ociImageConfig.RootFS.Type = "layers"
	ociImageConfig.History = []ociv1.History{}

	for i, layer := range converter.layers {
		timestamp, error := time.Parse(time.RFC3339, layer.History.Created)

		if error != nil {
			return hash, size, error
		}

		var cmd []string

		if layer.History.ContainerConfig != nil {
			cmd = layer.History.ContainerConfig.Cmd

		} else if layer.History.Config != nil {
			cmd = layer.History.Config.Cmd

		} else {
			return hash, size, fmt.Errorf("No history config found in layer %d of image '%s'", i, converter.imageName)
		}

		history := ociv1.History{Created: &timestamp, CreatedBy: strings.Join(cmd, " "), EmptyLayer: layer.EmptyLayer, Author: layer.History.Author}

		ociImageConfig.History = append(ociImageConfig.History, history)

		if layer.EmptyLayer {
			continue
		}

		digest, error := digest.Parse(layer.TarDigest)
		if error != nil {
			return hash, size, error
		}

		ociImageConfig.RootFS.DiffIDs = append(ociImageConfig.RootFS.DiffIDs, digest)
	}

	data, error := json.MarshalIndent(&ociImageConfig, "", "   ")
	if error != nil {
		return
	}

	sum := sha256.Sum256([]byte(data))

	hash = fmt.Sprintf("sha256:%x", sum)

	filename := path.Join("blobs", "sha256", hash[7:])

	error = converter.storage.WriteFile(filename, data)

	size = len(data)

	return
}

func (converter *imageConverter) writeOCIManifest(imageConfigHash string, imageConfigSize int) (hash string, size int, error error) {
	ociManifest := ociv1.Manifest{}
	ociManifest.SchemaVersion = 2
	ociManifest.Config.MediaType = "application/vnd.oci.image.config.v1+json"
	ociManifest.Config.Size = int64(imageConfigSize)

	ociManifest.Config.Digest, error = digest.Parse(imageConfigHash)
	if error != nil {
		return
	}

	ociManifest.Layers = []ociv1.Descriptor{}

	for _, layer := range converter.layers {
		if layer.EmptyLayer {
			continue
		}

		descriptor := ociv1.Descriptor{}

		descriptor.MediaType = layer.MediaType
		descriptor.Size = int64(layer.Size)

		descriptor.Digest, error = digest.Parse(layer.BlobDigest)
		if error != nil {
			return hash, size, error
		}

		ociManifest.Layers = append(ociManifest.Layers, descriptor)
	}

	data, error := json.MarshalIndent(&ociManifest, "", "   ")
	if error != nil {
		return
	}

	sum := sha256.Sum256([]byte(data))

	hash = fmt.Sprintf("sha256:%x", sum)

	filename := path.Join("blobs", "sha256", hash[7:])

	error = converter.storage.WriteFile(filename, data)

	size = len(data)

	return
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
