package download

import (
	"archive/tar"
	"compress/gzip"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/cavaliercoder/grab"
	"github.com/darxkies/k8s-tew/pkg/config"
	"github.com/darxkies/k8s-tew/pkg/container/image/converter"
	"github.com/darxkies/k8s-tew/pkg/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// CompressedFile contains the source name and target name of the a compressed file
type CompressedFile struct {
	SourceFile string
	TargetFile string
}

// Downloader contains data to download dependencies
type Downloader struct {
	config          *config.InternalConfig
	downloaderSteps utils.Tasks
	forceDownload   bool
	parallel        bool
	pullImages      bool
}

// NewDownloader creates a new instance of the downloader
func NewDownloader(config *config.InternalConfig, forceDownload bool, parallel bool, pullImages bool) Downloader {
	downloader := Downloader{config: config, forceDownload: forceDownload, parallel: parallel}

	downloader.downloaderSteps = utils.Tasks{}
	downloader.addTask(downloader.copyK8STEW)
	downloader.addTask(downloader.downloadEtcdBinaries)
	downloader.addTask(downloader.downloadKubernetesBinaries)
	downloader.addTask(downloader.downloadHelmBinary)
	downloader.addTask(downloader.downloadContainerdBinaries)
	downloader.addTask(downloader.downloadRuncBinary)
	downloader.addTask(downloader.downloadCriCtlBinary)
	downloader.addTask(downloader.downloadVeleroBinaries)

	if pullImages {
		downloader.addTask(downloader.downloadImages)
	}

	return downloader
}

func (downloader *Downloader) addTask(task utils.Task) {
	downloader.downloaderSteps = append(downloader.downloaderSteps, func() error {
		defer utils.IncreaseProgressStep()

		return task()
	})
}

// Steps returns the name of steps to be performed for the progress
func (downloader Downloader) Steps() int {
	result := len(downloader.downloaderSteps)

	if downloader.pullImages {
		// Images to download
		result += len(downloader.config.Config.Versions.GetImages())
	}

	return result
}

func (downloader Downloader) getURL(url, filename string) (string, error) {
	data := struct {
		Filename string
		Versions config.Versions
	}{
		Filename: filename,
		Versions: downloader.config.Config.Versions,
	}

	return utils.ApplyTemplate(url, url, data, false)
}

func (downloader Downloader) downloadFile(url, filename string) error {
	// Remove file to be downloaded
	os.Remove(filename)

	utils.LogURL("Downloading", url)

	// Create client
	client := grab.NewClient()

	// Set User Agent
	client.UserAgent = "k8s-tew"

	// Set connection timeout
	client.HTTPClient.Timeout = 10 * time.Second

	// Disable any proxies
	client.HTTPClient = &http.Client{
		Transport: &http.Transport{
			Proxy:           nil,
			TLSClientConfig: &tls.Config{},
		},
	}

	// Create new request
	request, error := grab.NewRequest(filename, url)
	if error != nil {
		return fmt.Errorf("Could not create request to download file %s from %s (%s)", filename, url, error.Error())
	}

	// Send request
	response := client.Do(request)

	// Check error
	if error := response.Err(); error != nil {
		return fmt.Errorf("Could not download file %s from %s (%s)", filename, url, error.Error())
	}

	return nil
}

func (downloader Downloader) downloadExecutable(urlTemplate, remoteFilename, filename string) error {
	url, error := downloader.getURL(urlTemplate, remoteFilename)
	if error != nil {
		return error
	}

	if !downloader.forceDownload && utils.FileExists(filename) {
		utils.LogURL("Skipped downloading", url)
		utils.LogFilename("Skipped installing", filename)

		return nil
	}

	temporaryFilename := path.Join(downloader.config.GetFullLocalAssetDirectory(utils.DirectoryTemporary), path.Base(filename))

	// Make sure the file is deleted once done
	defer func() {
		_ = os.Remove(temporaryFilename)
	}()

	if error := downloader.downloadFile(url, temporaryFilename); error != nil {
		return error
	}

	// Move target temporary file to target file
	if error := utils.MoveFile(temporaryFilename, filename); error != nil {
		return error
	}

	// Make target file executable
	if error := os.Chmod(filename, 0777); error != nil {
		return error
	}

	utils.LogFilename("Installed", filename)

	return nil
}

func (downloader Downloader) extractTGZ(filename string, targetDirectory string, compressed bool) error {
	// Remove any previous content
	os.RemoveAll(targetDirectory)

	// Create directory
	if error := utils.CreateDirectoryIfMissing(targetDirectory); error != nil {
		return errors.Wrapf(error, "could not create directory '%s'", targetDirectory)
	}

	// Open compressed file
	file, error := os.Open(filename)
	if error != nil {
		return errors.Wrapf(error, "could not open file '%s'", filename)
	}

	// Defer file close operation
	defer file.Close()

	var tarReader *tar.Reader

	if compressed {
		// Open gzip reader
		gzipReader, error := gzip.NewReader(file)
		if error != nil {
			return errors.Wrapf(error, "could not create gzip reader for '%s'", filename)
		}

		// Open tar reader
		tarReader = tar.NewReader(gzipReader)

	} else {
		// Open tar reader
		tarReader = tar.NewReader(file)
	}

	for {
		// Get tar header
		header, error := tarReader.Next()

		// Exit on end of file
		if error == io.EOF {
			break
		}

		// Exit if any other error occurred
		if error != nil {
			return errors.Wrapf(error, "could not get tar header from '%s'", filename)
		}

		fullName := path.Join(targetDirectory, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if error := utils.CreateDirectoryIfMissing(fullName); error != nil {
				return errors.Wrapf(error, "could not create directory '%s'", fullName)
			}

		case tar.TypeReg:
			log.WithFields(log.Fields{"filename": filepath.Dir(fullName)}).Debug("Extracting")

			if error := utils.CreateDirectoryIfMissing(filepath.Dir(fullName)); error != nil {
				return errors.Wrapf(error, "could not create directory '%s'", fullName)
			}

			outputFile, error := os.OpenFile(fullName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)

			if error != nil {
				return errors.Wrapf(error, "could not open file '%s'", fullName)
			}

			defer outputFile.Close()

			if _, error := io.Copy(outputFile, tarReader); error != nil {
				return errors.Wrapf(error, "could not write to file '%s'", fullName)
			}

		default:
		}
	}

	return nil
}

func (downloader Downloader) downloadAndExtractTGZFiles(urlTemplate, baseName string, files []CompressedFile, compressed bool) error {
	// Check if files already exist
	exist := true
	temporaryDirectory := downloader.config.GetFullLocalAssetDirectory(utils.DirectoryTemporary)

	if downloader.forceDownload {
		exist = false

	} else {
		for _, compressedFile := range files {
			if !utils.FileExists(compressedFile.TargetFile) {
				exist = false

				break
			}
		}
	}

	// Build base name including the version number
	baseName, error := downloader.getURL(baseName, "")
	if error != nil {
		return errors.Wrapf(error, "could not get url '%s'", baseName)
	}

	url, error := downloader.getURL(urlTemplate, baseName)
	if error != nil {
		return errors.Wrapf(error, "could not get url '%s'", baseName)
	}

	// All files exist, print skip message and bail out
	if exist {
		utils.LogURL("Skipped downloading", url)

		for _, compressedFile := range files {
			utils.LogFilename("Skipped installing", compressedFile.TargetFile)
		}

		return nil
	}

	// Create temporary download filename
	temporaryFile := path.Join(temporaryDirectory, baseName+".tgz")

	// Download file
	if error = downloader.downloadFile(url, temporaryFile); error != nil {
		return errors.Wrapf(error, "could not download file '%s'", url)
	}

	// Make sure the file is deleted once done
	defer func() {
		_ = os.Remove(temporaryFile)
	}()

	// Create temporary directory to extract to
	temporaryExtractedDirectory := path.Join(temporaryDirectory, baseName)

	// Extrace files
	if error := downloader.extractTGZ(temporaryFile, temporaryExtractedDirectory, compressed); error != nil {
		return errors.Wrapf(error, "could not extract tgz '%s' to '%s'", temporaryFile, temporaryExtractedDirectory)
	}

	// Make sure the temporary directory is removed once done
	defer func() {
		_ = os.RemoveAll(temporaryExtractedDirectory)
	}()

	// Move files from temporary directory to target directory
	for _, compressedFile := range files {
		sourceFilename := path.Join(temporaryExtractedDirectory, compressedFile.SourceFile)

		if error := utils.MoveFile(sourceFilename, compressedFile.TargetFile); error != nil {
			return errors.Wrapf(error, "could not rename '%s' to '%s'", sourceFilename, compressedFile.TargetFile)
		}

		utils.LogFilename("Installed", compressedFile.TargetFile)
	}

	return nil
}

func (downloader Downloader) copyK8STEW() error {
	binaryName, error := os.Executable()

	if error != nil {
		return error
	}

	targetFilename := downloader.config.GetFullLocalAssetFilename(utils.BinaryK8sTew)

	if binaryName == targetFilename {
		utils.LogFilename("Skipped", targetFilename)

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

	utils.LogFilename("Copied", targetFilename)

	return targetFile.Sync()
}

func (downloader Downloader) downloadHelmBinary() error {
	compressedFiles := []CompressedFile{
		{
			SourceFile: path.Join("linux-amd64", utils.BinaryHelm),
			TargetFile: downloader.config.GetFullLocalAssetFilename(utils.BinaryHelm),
		},
	}

	return downloader.downloadAndExtractTGZFiles(utils.HelmDownloadUrl, utils.HelmBaseName, compressedFiles, true)
}

func (downloader Downloader) downloadRuncBinary() error {
	return downloader.downloadExecutable(utils.RuncDownloadUrl, "", downloader.config.GetFullLocalAssetFilename(utils.BinaryRunc))
}

func (downloader Downloader) downloadEtcdBinaries() error {
	// Build base name including the version number
	baseName, error := downloader.getURL(utils.EtcdBaseName, "")
	if error != nil {
		return error
	}

	compressedFiles := []CompressedFile{
		{
			SourceFile: path.Join(baseName, utils.BinaryEtcdctl),
			TargetFile: downloader.config.GetFullLocalAssetFilename(utils.BinaryEtcdctl),
		},
	}

	return downloader.downloadAndExtractTGZFiles(utils.EtcdDownloadUrl, utils.EtcdBaseName, compressedFiles, false)
}

func (downloader Downloader) downloadKubernetesBinaries() error {
	kubernetesServerBin := path.Join("kubernetes", "node", "bin")

	compressedFiles := []CompressedFile{
		{
			SourceFile: path.Join(kubernetesServerBin, utils.BinaryKubelet),
			TargetFile: downloader.config.GetFullLocalAssetFilename(utils.BinaryKubelet),
		},
		{
			SourceFile: path.Join(kubernetesServerBin, utils.BinaryKubectl),
			TargetFile: downloader.config.GetFullLocalAssetFilename(utils.BinaryKubectl),
		},
	}

	return downloader.downloadAndExtractTGZFiles(utils.K8sDownloadUrl, utils.K8sBaseName, compressedFiles, true)
}

func (downloader Downloader) downloadContainerdBinaries() error {
	compressedFiles := []CompressedFile{
		{
			SourceFile: path.Join("bin", utils.BinaryContainerd),
			TargetFile: downloader.config.GetFullLocalAssetFilename(utils.BinaryContainerd),
		},
		{
			SourceFile: path.Join("bin", utils.BinaryContainerdShimRuncV2),
			TargetFile: downloader.config.GetFullLocalAssetFilename(utils.BinaryContainerdShimRuncV2),
		},
		{
			SourceFile: path.Join("bin", utils.BinaryCtr),
			TargetFile: downloader.config.GetFullLocalAssetFilename(utils.BinaryCtr),
		},
	}

	return downloader.downloadAndExtractTGZFiles(utils.ContainerdDownloadUrl, utils.ContainerdBaseName, compressedFiles, true)
}

func (downloader Downloader) downloadCriCtlBinary() error {
	compressedFiles := []CompressedFile{
		{
			SourceFile: utils.BinaryCrictl,
			TargetFile: downloader.config.GetFullLocalAssetFilename(utils.BinaryCrictl),
		},
	}

	return downloader.downloadAndExtractTGZFiles(utils.CrictlDownloadUrl, utils.CrictlBaseName, compressedFiles, true)
}

func (downloader Downloader) downloadVeleroBinaries() error {
	// Build base name including the version number
	baseName, error := downloader.getURL(utils.VeleroBaseName, "")
	if error != nil {
		return error
	}

	compressedFiles := []CompressedFile{
		{
			SourceFile: path.Join(baseName, utils.BinaryVelero),
			TargetFile: downloader.config.GetFullLocalAssetFilename(utils.BinaryVelero),
		},
	}

	return downloader.downloadAndExtractTGZFiles(utils.VeleroDownloadUrl, utils.VeleroBaseName, compressedFiles, true)
}

func (downloader Downloader) downloadImages() error {
	for _, image := range downloader.config.Config.Versions.GetImages() {
		imageFilename := downloader.config.GetFullLocalAssetFilename(image.GetImageFilename())

		if utils.FileExists(imageFilename) {
			log.WithFields(log.Fields{"image": image.Name, "_filename": imageFilename}).Debug("Skipped downloading")

			utils.IncreaseProgressStep()

			continue
		}

		log.WithFields(log.Fields{"name": image.Name, "_filename": imageFilename}).Info("Downloading image")

		error := converter.PullImage(image.Name, imageFilename, false)
		if error != nil {
			return error
		}

		log.WithFields(log.Fields{"name": image.Name, "_filename": imageFilename}).Info("Downloaded image")

		utils.IncreaseProgressStep()
	}

	utils.IncreaseProgressStep()

	return nil
}

func (downloader Downloader) createLocalDirectories() error {
	for name, directory := range downloader.config.Config.Assets.Directories {
		if directory.Absolute {
			continue
		}

		localDirectory := downloader.config.GetFullLocalAssetDirectory(name)

		if error := utils.CreateDirectoryIfMissing(localDirectory); error != nil {
			return error
		}
	}

	return nil
}

// DownloadBinaries downloads all required binaries
func (downloader Downloader) DownloadBinaries() error {
	if error := downloader.createLocalDirectories(); error != nil {
		return error
	}

	errors := utils.RunParallelTasks(downloader.downloaderSteps, downloader.parallel)
	if len(errors) > 0 {
		return errors[0]
	}

	return nil
}
