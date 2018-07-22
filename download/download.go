package download

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/darxkies/k8s-tew/config"
	"github.com/darxkies/k8s-tew/utils"

	log "github.com/sirupsen/logrus"
)

type CompressedFile struct {
	SourceFile string
	TargetFile string
}

type Downloader struct {
	config *config.InternalConfig
}

func NewDownloader(config *config.InternalConfig) Downloader {
	return Downloader{config: config}
}

func (downloader Downloader) getURL(url, filename string) (string, error) {
	data := struct {
		Filename string
		Versions config.Versions
	}{
		Filename: filename,
		Versions: downloader.config.Config.Versions,
	}

	return utils.ApplyTemplate(url, data)
}

func (downloader Downloader) downloadFile(urlTemplate, remoteFilename, filename string) (bool, error) {
	url, error := downloader.getURL(urlTemplate, remoteFilename)

	if error != nil {
		return false, error
	}

	if utils.FileExists(filename) {
		log.WithFields(log.Fields{"filename": filename}).Info("skipped")

		return false, nil
	}

	log.WithFields(log.Fields{"url": url}).Info("downloading")

	output, error := os.Create(filename)

	if error != nil {
		return false, error
	}

	defer output.Close()

	response, error := http.Get(url)

	if error != nil {
		return false, error
	}

	defer response.Body.Close()

	_, error = io.Copy(output, response.Body)

	return true, error
}

func (downloader Downloader) downloadExecutable(urlTemplate, remoteFilename, filename string) error {
	url, error := downloader.getURL(urlTemplate, remoteFilename)

	if error != nil {
		return error
	}

	installed, error := downloader.downloadFile(url, remoteFilename, filename)
	if error != nil {
		return error
	}

	if error := os.Chmod(filename, 0555); error != nil {
		return error
	}

	if installed {
		log.WithFields(log.Fields{"filename": filename}).Info("installed")
	}

	return nil
}

func (downloader Downloader) extractTGZ(filename string, targetDirectory string) error {
	if error := utils.CreateDirectoryIfMissing(targetDirectory); error != nil {
		return error
	}

	file, error := os.Open(filename)

	if error != nil {
		return error
	}

	defer file.Close()

	gzipReader, error := gzip.NewReader(file)

	if error != nil {
		return error
	}

	tarReader := tar.NewReader(gzipReader)

	for true {
		header, error := tarReader.Next()

		if error == io.EOF {
			break
		}

		if error != nil {
			return error
		}

		fullName := path.Join(targetDirectory, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if error := utils.CreateDirectoryIfMissing(fullName); error != nil {
				return error
			}

		case tar.TypeReg:
			outputFile, error := os.OpenFile(fullName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0555)

			if error != nil {
				return error
			}

			defer outputFile.Close()

			if _, error := io.Copy(outputFile, tarReader); error != nil {
				return error
			}
		default:
		}
	}

	return nil
}

func (downloader Downloader) downloadAndExtractTGZFiles(urlTemplate, baseName string, files []CompressedFile) error {
	// Check if files already exist
	exist := true
	temporaryDirectory := downloader.config.GetFullLocalAssetDirectory(utils.TEMPORARY_DIRECTORY)

	for _, compressedFile := range files {
		if !utils.FileExists(compressedFile.TargetFile) {
			exist = false

			break
		}
	}

	// All files exist, print skip message and bail out
	if exist {
		for _, compressedFile := range files {
			log.WithFields(log.Fields{"filename": compressedFile.TargetFile}).Info("skipped")
		}

		return nil
	}

	// Build base name including the version number
	baseName, error := downloader.getURL(baseName, "")
	if error != nil {
		return error
	}

	// Create temporary download filename
	temporaryFile := path.Join(temporaryDirectory, baseName+".tgz")

	// Download file
	_, error = downloader.downloadFile(urlTemplate, baseName, temporaryFile)
	if error != nil {
		return error
	}

	// Make sure the file is deleted once done
	defer func() {
		_ = os.Remove(temporaryFile)
	}()

	// Create temporary directory to extract to
	temporaryExtractedDirectory := path.Join(temporaryDirectory, baseName)

	// Extrace files
	if error := downloader.extractTGZ(temporaryFile, temporaryExtractedDirectory); error != nil {
		return error
	}

	// Make sure the temporary directory is removed once done
	defer func() {
		_ = os.RemoveAll(temporaryExtractedDirectory)
	}()

	// Move files from temporary directory to target directory
	for _, compressedFile := range files {
		if error := os.Rename(path.Join(temporaryExtractedDirectory, compressedFile.SourceFile), compressedFile.TargetFile); error != nil {
			return error
		}

		log.WithFields(log.Fields{"filename": compressedFile.TargetFile}).Info("installed")
	}

	return nil
}

func (downloader Downloader) copyK8STEW() error {
	binaryName, error := os.Executable()

	if error != nil {
		return error
	}

	targetFilename := downloader.config.GetFullLocalAssetFilename(utils.K8S_TEW_BINARY)

	if binaryName == targetFilename {
		log.WithFields(log.Fields{"filename": targetFilename}).Info("skipped")

		return nil
	}

	sourceFile, error := os.Open(binaryName)

	if error != nil {
		return error
	}

	defer sourceFile.Close()

	targetFile, error := os.OpenFile(targetFilename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)

	if error != nil {
		return error
	}

	defer targetFile.Close()

	_, error = io.Copy(targetFile, sourceFile)

	if error != nil {
		return error
	}

	log.WithFields(log.Fields{"filename": targetFilename}).Info("copied")

	return targetFile.Sync()
}

func (downloader Downloader) downloadK8SBinaries() error {
	if error := downloader.downloadExecutable(utils.K8S_DOWNLOAD_URL, utils.KUBECTL_BINARY, downloader.config.GetFullLocalAssetFilename(utils.KUBECTL_BINARY)); error != nil {
		return error
	}

	if error := downloader.downloadExecutable(utils.K8S_DOWNLOAD_URL, utils.KUBE_APISERVER_BINARY, downloader.config.GetFullLocalAssetFilename(utils.KUBE_APISERVER_BINARY)); error != nil {
		return error
	}

	if error := downloader.downloadExecutable(utils.K8S_DOWNLOAD_URL, utils.KUBE_CONTROLLER_MANAGER_BINARY, downloader.config.GetFullLocalAssetFilename(utils.KUBE_CONTROLLER_MANAGER_BINARY)); error != nil {
		return error
	}

	if error := downloader.downloadExecutable(utils.K8S_DOWNLOAD_URL, utils.KUBE_SCHEDULER_BINARY, downloader.config.GetFullLocalAssetFilename(utils.KUBE_SCHEDULER_BINARY)); error != nil {
		return error
	}

	if error := downloader.downloadExecutable(utils.K8S_DOWNLOAD_URL, utils.KUBE_PROXY_BINARY, downloader.config.GetFullLocalAssetFilename(utils.KUBE_PROXY_BINARY)); error != nil {
		return error
	}

	if error := downloader.downloadExecutable(utils.K8S_DOWNLOAD_URL, utils.KUBELET_BINARY, downloader.config.GetFullLocalAssetFilename(utils.KUBELET_BINARY)); error != nil {
		return error
	}

	return nil
}

func (downloader Downloader) downloadHelmBinary() error {
	compressedFiles := []CompressedFile{
		CompressedFile{
			SourceFile: path.Join("linux-amd64", utils.HELM_BINARY),
			TargetFile: downloader.config.GetFullLocalAssetFilename(utils.HELM_BINARY),
		},
	}

	return downloader.downloadAndExtractTGZFiles(utils.HELM_DOWNLOAD_URL, utils.HELM_BASE_NAME, compressedFiles)
}

func (downloader Downloader) downloadRuncBinary() error {
	return downloader.downloadExecutable(utils.RUNC_DOWNLOAD_URL, "", downloader.config.GetFullLocalAssetFilename(utils.RUNC_BINARY))
}

func (downloader Downloader) downloadFlanneldBinary() error {
	return downloader.downloadExecutable(utils.FLANNELD_DOWNLOAD_URL, "", downloader.config.GetFullLocalAssetFilename(utils.FLANNELD_BINARY))
}

func (downloader Downloader) downloadEtcdBinaries() error {
	// Build base name including the version number
	baseName, error := downloader.getURL(utils.ETCD_BASE_NAME, "")
	if error != nil {
		return error
	}

	compressedFiles := []CompressedFile{
		CompressedFile{
			SourceFile: path.Join(baseName, utils.ETCD_BINARY),
			TargetFile: downloader.config.GetFullLocalAssetFilename(utils.ETCD_BINARY),
		},
		CompressedFile{
			SourceFile: path.Join(baseName, utils.ETCDCTL_BINARY),
			TargetFile: downloader.config.GetFullLocalAssetFilename(utils.ETCDCTL_BINARY),
		},
	}

	return downloader.downloadAndExtractTGZFiles(utils.ETCD_DOWNLOAD_URL, utils.ETCD_BASE_NAME, compressedFiles)
}

func (downloader Downloader) downloadCNIBinaries() error {
	compressedFiles := []CompressedFile{
		CompressedFile{
			SourceFile: utils.BRIDGE_BINARY,
			TargetFile: downloader.config.GetFullLocalAssetFilename(utils.BRIDGE_BINARY),
		},
		CompressedFile{
			SourceFile: utils.FLANNEL_BINARY,
			TargetFile: downloader.config.GetFullLocalAssetFilename(utils.FLANNEL_BINARY),
		},
		CompressedFile{
			SourceFile: utils.LOOPBACK_BINARY,
			TargetFile: downloader.config.GetFullLocalAssetFilename(utils.LOOPBACK_BINARY),
		},
		CompressedFile{
			SourceFile: utils.HOST_LOCAL_BINARY,
			TargetFile: downloader.config.GetFullLocalAssetFilename(utils.HOST_LOCAL_BINARY),
		},
	}

	return downloader.downloadAndExtractTGZFiles(utils.CNI_DOWNLOAD_URL, utils.CNI_BASE_NAME, compressedFiles)
}

func (downloader Downloader) downloadContainerdBinaries() error {
	compressedFiles := []CompressedFile{
		CompressedFile{
			SourceFile: path.Join("bin", utils.CONTAINERD_BINARY),
			TargetFile: downloader.config.GetFullLocalAssetFilename(utils.CONTAINERD_BINARY),
		},
		CompressedFile{
			SourceFile: path.Join("bin", utils.CONTAINERD_SHIM_BINARY),
			TargetFile: downloader.config.GetFullLocalAssetFilename(utils.CONTAINERD_SHIM_BINARY),
		},
		CompressedFile{
			SourceFile: path.Join("bin", utils.CTR_BINARY),
			TargetFile: downloader.config.GetFullLocalAssetFilename(utils.CTR_BINARY),
		},
	}

	return downloader.downloadAndExtractTGZFiles(utils.CONTAINERD_DOWNLOAD_URL, utils.CONTAINERD_BASE_NAME, compressedFiles)
}

func (downloader Downloader) downloadCriCtlBinary() error {
	compressedFiles := []CompressedFile{
		CompressedFile{
			SourceFile: utils.CRICTL_BINARY,
			TargetFile: downloader.config.GetFullLocalAssetFilename(utils.CRICTL_BINARY),
		},
	}

	return downloader.downloadAndExtractTGZFiles(utils.CRICTL_DOWNLOAD_URL, utils.CRICTL_BASE_NAME, compressedFiles)
}

func (downloader Downloader) downloadGobetweenBinary() error {
	compressedFiles := []CompressedFile{
		CompressedFile{
			SourceFile: utils.GOBETWEEN_BINARY,
			TargetFile: downloader.config.GetFullLocalAssetFilename(utils.GOBETWEEN_BINARY),
		},
	}

	return downloader.downloadAndExtractTGZFiles(utils.GOBETWEEN_DOWNLOAD_URL, utils.GOBETWEEN_BASE_NAME, compressedFiles)
}

func (downloader Downloader) createLocalDirectories() error {
	for name := range downloader.config.Config.Assets.Directories {
		localDirectory := downloader.config.GetFullLocalAssetDirectory(name)

		if error := utils.CreateDirectoryIfMissing(localDirectory); error != nil {
			return error
		}
	}

	return nil
}

func (downloader Downloader) DownloadBinaries() error {
	if error := downloader.createLocalDirectories(); error != nil {
		return error
	}

	if error := downloader.copyK8STEW(); error != nil {
		return error
	}

	if error := downloader.downloadEtcdBinaries(); error != nil {
		return error
	}

	if error := downloader.downloadFlanneldBinary(); error != nil {
		return error
	}

	if error := downloader.downloadK8SBinaries(); error != nil {
		return error
	}

	if error := downloader.downloadHelmBinary(); error != nil {
		return error
	}

	if error := downloader.downloadCNIBinaries(); error != nil {
		return error
	}

	if error := downloader.downloadContainerdBinaries(); error != nil {
		return error
	}

	if error := downloader.downloadRuncBinary(); error != nil {
		return error
	}

	if error := downloader.downloadCriCtlBinary(); error != nil {
		return error
	}

	if error := downloader.downloadGobetweenBinary(); error != nil {
		return error
	}

	return nil
}
