package download

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/darxkies/k8s-tew/config"
	"github.com/darxkies/k8s-tew/utils"

	log "github.com/sirupsen/logrus"
)

type Downloader struct {
	config          *config.InternalConfig
	etcdVersion     string
	flanneldVersion string
	k8sVersion      string
	cniVersion      string
	criVersion      string
}

func NewDownloader(config *config.InternalConfig, etcdVersion string, flanneldVersion string, k8sVersion string, cniVersion string, criVersion string) Downloader {
	return Downloader{config: config, etcdVersion: etcdVersion, flanneldVersion: flanneldVersion, k8sVersion: k8sVersion, cniVersion: cniVersion, criVersion: criVersion}
}

func (downloader Downloader) downloadFile(url, filename string) error {
	if utils.FileExists(filename) {
		log.WithFields(log.Fields{"filename": filename}).Info("skipping")

		return nil
	}

	log.WithFields(log.Fields{"url": url}).Info("downloading")

	output, error := os.Create(filename)

	if error != nil {
		return error
	}

	defer output.Close()

	response, error := http.Get(url)

	if error != nil {
		return error
	}

	defer response.Body.Close()

	_, error = io.Copy(output, response.Body)

	if error == nil {
		log.WithFields(log.Fields{"filename": filename}).Info("downloaded")
	}

	return error
}

func (downloader Downloader) downloadExecutable(url, filename string) error {
	if error := downloader.downloadFile(url, filename); error != nil {
		return error
	}

	if error := os.Chmod(filename, 0555); error != nil {
		return error
	}

	return nil
}

func (downloader Downloader) extractTGZ(filename string, targetDirectory string) error {
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

func (downloader Downloader) downloadEtcdBinaries() error {
	etcdFullName := downloader.config.GetFullDeploymentFilename(utils.ETCD_BINARY)
	etcdctlFullName := downloader.config.GetFullDeploymentFilename(utils.ETCDCTL_BINARY)

	if utils.FileExists(etcdFullName) && utils.FileExists(etcdctlFullName) {
		log.WithFields(log.Fields{"filename": etcdFullName}).Info("skipping")
		log.WithFields(log.Fields{"filename": etcdctlFullName}).Info("skipping")

		return nil
	}

	temporaryFile := path.Join(utils.GetFullTemporaryDirectory(), "etcd.tgz")

	baseName := fmt.Sprintf(utils.ETCD_BASE_NAME, downloader.etcdVersion)

	if error := downloader.downloadFile(fmt.Sprintf(utils.ETCD_DOWNLOAD_URL, downloader.etcdVersion, baseName), temporaryFile); error != nil {
		return error
	}

	defer func() {
		_ = os.Remove(temporaryFile)
	}()

	if error := downloader.extractTGZ(temporaryFile, utils.GetFullTemporaryDirectory()); error != nil {
		return error
	}

	extractedDirectoryName := path.Join(utils.GetFullTemporaryDirectory(), baseName)

	defer func() {
		_ = os.RemoveAll(extractedDirectoryName)
	}()

	if error := os.Rename(path.Join(extractedDirectoryName, utils.ETCD_BINARY), etcdFullName); error != nil {
		return nil
	}

	if error := os.Rename(path.Join(extractedDirectoryName, utils.ETCDCTL_BINARY), etcdctlFullName); error != nil {
		return nil
	}

	log.WithFields(log.Fields{"filename": etcdFullName}).Info("downloaded")
	log.WithFields(log.Fields{"filename": etcdctlFullName}).Info("downloaded")

	return nil
}

func (downloader Downloader) downloadFlanneldBinary() error {
	flanneldFullName := downloader.config.GetFullDeploymentFilename(utils.FLANNELD_BINARY)

	if utils.FileExists(flanneldFullName) {
		log.WithFields(log.Fields{"filename": flanneldFullName}).Info("skipping")

		return nil
	}

	if error := downloader.downloadExecutable(fmt.Sprintf(utils.FLANNELD_DOWNLOAD_URL, downloader.flanneldVersion), downloader.config.GetFullDeploymentFilename(utils.FLANNELD_BINARY)); error != nil {
		return nil
	}

	log.WithFields(log.Fields{"filename": flanneldFullName}).Info("downloaded")

	return nil
}

func (downloader Downloader) downloadCNIBinaries() error {
	cniDirectory := path.Join(downloader.config.BaseDirectory, utils.GetFullCNIBinariesDirectory())

	bridgeFullName := downloader.config.GetFullDeploymentFilename(utils.BRIDGE_BINARY)

	if utils.FileExists(bridgeFullName) {
		log.WithFields(log.Fields{"filename": bridgeFullName}).Info("skipping")

		return nil
	}

	temporaryFile := path.Join(utils.GetFullTemporaryDirectory(), "cni.tgz")

	baseName := fmt.Sprintf(utils.CNI_BASE_NAME, downloader.cniVersion)

	if error := downloader.downloadFile(fmt.Sprintf(utils.CNI_DOWNLOAD_URL, downloader.cniVersion, baseName), temporaryFile); error != nil {
		return error
	}

	defer func() {
		_ = os.Remove(temporaryFile)
	}()

	if error := downloader.extractTGZ(temporaryFile, cniDirectory); error != nil {
		return error
	}

	log.WithFields(log.Fields{"filename": bridgeFullName}).Info("downloaded")

	return nil
}

func (downloader Downloader) downloadCRIBinaries() error {
	criDirectory := path.Join(downloader.config.BaseDirectory, utils.GetFullCRIBinariesDirectory())

	runcFullName := downloader.config.GetFullDeploymentFilename(utils.RUNC_BINARY)

	if utils.FileExists(runcFullName) {
		log.WithFields(log.Fields{"filename": runcFullName}).Info("skipping")

		return nil
	}

	temporaryBaseFilename := "cri"
	temporaryFile := path.Join(utils.GetFullTemporaryDirectory(), temporaryBaseFilename+".tgz")

	if error := downloader.downloadFile(fmt.Sprintf(utils.CRI_DOWNLOAD_URL, downloader.criVersion, downloader.criVersion), temporaryFile); error != nil {
		return error
	}

	defer func() {
		_ = os.Remove(temporaryFile)
	}()

	temporaryCRIDirectory := path.Join(utils.GetFullTemporaryDirectory(), temporaryBaseFilename)

	if error := downloader.extractTGZ(temporaryFile, temporaryCRIDirectory); error != nil {
		return error
	}

	defer func() {
		_ = os.RemoveAll(temporaryCRIDirectory)
	}()

	if error := os.Rename(path.Join(temporaryCRIDirectory, "usr", "local", "sbin", "runc"), runcFullName); error != nil {
		return nil
	}

	for _, binary := range []string{utils.CONTAINERD_BINARY, utils.CRI_CONTAINERD_BINARY, utils.CONTAINERD_SHIM_BINARY, utils.CTR_BINARY, utils.CRICTL_BINARY} {
		if error := os.Rename(path.Join(temporaryCRIDirectory, "usr", "local", "bin", binary), path.Join(criDirectory, binary)); error != nil {
			return nil
		}
	}

	return nil
}

func (downloader Downloader) downloadK8SBinaries() error {
	if error := downloader.downloadExecutable(fmt.Sprintf(utils.K8S_DOWNLOAD_URL, downloader.k8sVersion, utils.KUBECTL_BINARY), downloader.config.GetFullDeploymentFilename(utils.KUBECTL_BINARY)); error != nil {
		return error
	}

	if error := downloader.downloadExecutable(fmt.Sprintf(utils.K8S_DOWNLOAD_URL, downloader.k8sVersion, utils.KUBE_APISERVER_BINARY), downloader.config.GetFullDeploymentFilename(utils.KUBE_APISERVER_BINARY)); error != nil {
		return error
	}

	if error := downloader.downloadExecutable(fmt.Sprintf(utils.K8S_DOWNLOAD_URL, downloader.k8sVersion, utils.KUBE_CONTROLLER_MANAGER_BINARY), downloader.config.GetFullDeploymentFilename(utils.KUBE_CONTROLLER_MANAGER_BINARY)); error != nil {
		return error
	}

	if error := downloader.downloadExecutable(fmt.Sprintf(utils.K8S_DOWNLOAD_URL, downloader.k8sVersion, utils.KUBE_SCHEDULER_BINARY), downloader.config.GetFullDeploymentFilename(utils.KUBE_SCHEDULER_BINARY)); error != nil {
		return error
	}

	if error := downloader.downloadExecutable(fmt.Sprintf(utils.K8S_DOWNLOAD_URL, downloader.k8sVersion, utils.KUBE_PROXY_BINARY), downloader.config.GetFullDeploymentFilename(utils.KUBE_PROXY_BINARY)); error != nil {
		return error
	}

	if error := downloader.downloadExecutable(fmt.Sprintf(utils.K8S_DOWNLOAD_URL, downloader.k8sVersion, utils.KUBELET_BINARY), downloader.config.GetFullDeploymentFilename(utils.KUBELET_BINARY)); error != nil {
		return error
	}

	return nil
}

func (downloader Downloader) copyK8STEW() error {
	binaryName, error := os.Executable()

	if error != nil {
		return error
	}

	targetFilename := downloader.config.GetFullDeploymentFilename(utils.K8S_TEW_BINARY)

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

	log.WithFields(log.Fields{"filename": utils.K8S_TEW_BINARY}).Info("copied")

	return targetFile.Sync()
}

func (downloader Downloader) DownloadBinaries() error {
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

	if error := downloader.downloadCNIBinaries(); error != nil {
		return error
	}

	if error := downloader.downloadCRIBinaries(); error != nil {
		return error
	}

	return nil
}
