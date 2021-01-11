package deployment

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/darxkies/k8s-tew/pkg/config"
	"github.com/darxkies/k8s-tew/pkg/k8s"
	"github.com/darxkies/k8s-tew/pkg/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/tmc/scp"
	"golang.org/x/crypto/ssh"
)

type NodeDeployment struct {
	identityFile            string
	name                    string
	node                    *config.Node
	config                  *config.InternalConfig
	sshLimiter              *utils.Limiter
	parallel                bool
	targetChecksumsFilename string
	localChecksums          *utils.Checksums
}

func NewNodeDeployment(identityFile string, name string, node *config.Node, config *config.InternalConfig, parallel bool) *NodeDeployment {
	return &NodeDeployment{identityFile: identityFile, name: name, node: node, config: config, sshLimiter: utils.NewLimiter(utils.ConcurrentSshConnectionsLimit), parallel: parallel, targetChecksumsFilename: path.Join(config.GetFullTargetAssetDirectory(utils.DirectoryDynamicData), "checksums"), localChecksums: utils.NewChecksums(path.Join(config.GetFullLocalAssetDirectory(utils.DirectoryDynamicData), "checksums"), config.BaseDirectory)}
}

func (deployment *NodeDeployment) Steps(skipRestart bool) (result int) {
	result = 0

	// Create Directories
	result++

	if !skipRestart {
		// Stop service
		result++
	}

	// Upload files
	result += len(deployment.config.Config.Assets.Files)

	// Cleanup workers
	result++

	if !skipRestart {
		// Start service
		result++
	}

	// Update checksums
	result++

	return
}

func (deployment *NodeDeployment) md5sum(filename string) (result string, error error) {
	return deployment.localChecksums.GetChecksum(filename)
}

func (deployment *NodeDeployment) createDirectories() error {
	defer utils.IncreaseProgressStep()

	directories := map[string]bool{}

	// Collect remote directories based on the files that have to be uploaded
	for _, file := range deployment.config.Config.Assets.Files {
		if !config.CompareLabels(deployment.node.Labels, file.Labels) {
			continue
		}

		directories[deployment.config.GetFullTargetAssetDirectory(file.Directory)] = true
	}

	// Collect remote directories based on their labels
	for name, directory := range deployment.config.Config.Assets.Directories {
		if !config.CompareLabels(deployment.node.Labels, directory.Labels) {
			continue
		}

		directories[deployment.config.GetFullTargetAssetDirectory(name)] = true
	}

	if len(directories) == 0 {
		return nil
	}

	// Create remote directories
	createDirectoriesCommand := "mkdir -p"

	for directoryName := range directories {
		createDirectoriesCommand += " " + directoryName
	}

	if _, error := deployment.Execute("create-directories", createDirectoriesCommand); error != nil {
		return error
	}

	return nil
}

func (deployment *NodeDeployment) getFiles() map[string]string {
	files := map[string]string{}

	// Collect files to be deployed
	for name, file := range deployment.config.Config.Assets.Files {
		if !config.CompareLabels(deployment.node.Labels, file.Labels) {
			continue
		}

		fromFile := deployment.config.GetFullLocalAssetFilename(name)
		toFile := deployment.config.GetFullTargetAssetFilename(name)

		// Skip if the file does not exist locally
		if !utils.FileExists(fromFile) {
			continue
		}

		files[fromFile] = toFile
	}

	return files
}

func (deployment *NodeDeployment) getRemoteFileChecksums() map[string]string {
	// Calculate checksums of remote files
	checksumCommand := "for i in"

	for _, toFile := range deployment.getFiles() {
		checksumCommand += fmt.Sprintf(" '%s'", toFile)
	}

	checksumCommand += fmt.Sprintf("; do if [ -f $i ]; then if [ $i -ot '%s' ]; then grep -e \" $i$\" '%s'; else md5sum $i; fi; fi; done", deployment.targetChecksumsFilename, deployment.targetChecksumsFilename)

	output, _ := deployment.Execute("get-checksums", checksumCommand)

	// Parse remote checksum values
	checksums := map[string]string{}
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		tokens := strings.Split(line, " ")

		checksums[tokens[len(tokens)-1]] = tokens[0]
	}

	return checksums
}

func (deployment *NodeDeployment) getChangedFiles() map[string]string {
	remoteFileChecksums := deployment.getRemoteFileChecksums()

	files := map[string]string{}

	for fromFile, toFile := range deployment.getFiles() {
		// Skip if the file does not exist locally
		if !utils.FileExists(fromFile) {
			continue
		}

		if remoteChecksum, ok := remoteFileChecksums[toFile]; ok {
			localChecksum, error := deployment.md5sum(fromFile)

			if error == nil && localChecksum == remoteChecksum {
				continue
			}
		}

		files[fromFile] = toFile
	}

	return files
}

func (deployment *NodeDeployment) UploadFiles(forceUpload bool, skipRestart bool) (_error error) {
	if _error = deployment.createDirectories(); _error != nil {
		return
	}

	var files map[string]string

	if forceUpload {
		files = deployment.getFiles()
	} else {
		files = deployment.getChangedFiles()
	}

	if len(files) > 0 && skipRestart == false {
		// Stop service
		_, _ = deployment.Execute("stop-service", fmt.Sprintf("systemctl stop %s", utils.ServiceName))
	}

	utils.IncreaseProgressStep()

	tasks := utils.Tasks{}

	filesList := ""

	for name, file := range deployment.config.Config.Assets.Files {
		fromFile := deployment.config.GetFullLocalAssetFilename(name)
		toFile := deployment.config.GetFullTargetAssetFilename(name)

		if !config.CompareLabels(deployment.node.Labels, file.Labels) {
			utils.IncreaseProgressStep()

			continue
		}

		if _, ok := files[fromFile]; !ok {
			utils.IncreaseProgressStep()

			continue
		}

		if len(filesList) >= 0 {
			filesList += " "
		}

		filesList += fmt.Sprintf("'%s'", toFile)

		tasks = append(tasks, func() error {
			defer utils.IncreaseProgressStep()

			return deployment.UploadFile(fromFile, toFile)
		})
	}

	// Upload files
	if errors := utils.RunParallelTasks(tasks, deployment.parallel); len(errors) > 0 {
		return errors[0]
	}

	cleanupFiles := []string{}

	// Remove controller manifests on controllers
	if deployment.node.IsControllerOnly() {
		if len(deployment.config.Config.ControllerVirtualIP) == 0 || len(deployment.config.Config.ControllerVirtualIPInterface) == 0 {
			cleanupFiles = append(cleanupFiles, deployment.config.GetFullTargetAssetFilename(utils.ManifestControllerVirtualIP))
		}
	}

	// Remove controller manifests on workers
	if deployment.node.IsWorkerOnly() {
		cleanupFiles = append(cleanupFiles, deployment.config.GetFullTargetAssetFilename(utils.ManifestControllerVirtualIP))
		cleanupFiles = append(cleanupFiles, deployment.config.GetFullTargetAssetFilename(utils.ManifestEtcd))
		cleanupFiles = append(cleanupFiles, deployment.config.GetFullTargetAssetFilename(utils.ManifestKubeApiserver))
		cleanupFiles = append(cleanupFiles, deployment.config.GetFullTargetAssetFilename(utils.ManifestKubeControllerManager))
		cleanupFiles = append(cleanupFiles, deployment.config.GetFullTargetAssetFilename(utils.ManifestKubeScheduler))

		if len(deployment.config.Config.WorkerVirtualIP) == 0 || len(deployment.config.Config.WorkerVirtualIPInterface) == 0 {
			cleanupFiles = append(cleanupFiles, deployment.config.GetFullTargetAssetFilename(utils.ManifestWorkerVirtualIP))
		}
	}

	if len(cleanupFiles) > 0 {
		_, _error = deployment.Execute("cleanup-files", fmt.Sprintf("rm -Rf %s", strings.Join(cleanupFiles, " ")))

		if _error != nil {
			return _error
		}
	}

	utils.IncreaseProgressStep()

	if len(files) > 0 && skipRestart == false {
		// Registrate and start service
		_, _error = deployment.Execute("start-service", fmt.Sprintf("systemctl daemon-reload && systemctl enable %s && systemctl start %s", utils.ServiceName, utils.ServiceName))
	}

	utils.IncreaseProgressStep()

	if len(filesList) > 0 {
		command := fmt.Sprintf("for i in %s; do if [ -f '%s' ]; then grep -ve \" $i$\" '%s' > /tmp/checksums; mv /tmp/checksums %s; fi; md5sum $i >> '%s'; done", filesList, deployment.targetChecksumsFilename, deployment.targetChecksumsFilename, deployment.targetChecksumsFilename, deployment.targetChecksumsFilename)

		_, _error = deployment.Execute("update-checkups", command)

		if _error != nil {
			return _error
		}
	}

	utils.IncreaseProgressStep()

	return
}

func (deployment *NodeDeployment) getSession() (*ssh.Session, error) {
	privateKeyContent, error := ioutil.ReadFile(deployment.identityFile)
	if error != nil {
		return nil, error
	}

	privateKey, error := ssh.ParsePrivateKey(privateKeyContent)
	if error != nil {
		return nil, error
	}

	client, error := ssh.Dial("tcp", deployment.node.IP+":22", &ssh.ClientConfig{
		User: utils.DeploymentUser,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(privateKey),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	if error != nil {
		return nil, error
	}

	return client.NewSession()
}

func (deployment *NodeDeployment) pullImage(image string) error {
	deployment.sshLimiter.Lock()
	defer deployment.sshLimiter.Unlock()

	crictl := deployment.config.GetFullTargetAssetFilename(utils.BinaryCrictl)
	containerdSock := deployment.config.GetFullTargetAssetFilename(utils.ContainerdSock)
	command := fmt.Sprintf("CONTAINER_RUNTIME_ENDPOINT=unix://%s %s pull %s", containerdSock, crictl, image)

	if _, error := deployment.Execute(fmt.Sprintf("pull-image-%s", image), command); error != nil {
		return fmt.Errorf("Failed to pull image %s (%s)", image, error.Error())
	}

	return nil
}

func GetImportImageCommand(config *config.InternalConfig, image, filename string) string {
	tokens := strings.Split(image, ":")

	baseName := image

	// Remove tag
	if len(tokens) == 2 {
		baseName = tokens[0]
	}

	ctr := fmt.Sprintf("CONTAINERD_NAMESPACE=\"%s\" \"%s\"", utils.ContainerdKubernetesNamespace, config.GetFullTargetAssetFilename(utils.BinaryCtr))

	command := fmt.Sprintf("%s i ls -q | grep -e \"^%s$\"; [ $? -eq 0 ] || %s i import --digests --base-name %s %s", ctr, image, ctr, baseName, filename)

	return command
}

func (deployment *NodeDeployment) importImage(image string, filename string) error {
	deployment.sshLimiter.Lock()
	defer deployment.sshLimiter.Unlock()

	command := GetImportImageCommand(deployment.config, image, filename)

	if _, error := deployment.Execute(fmt.Sprintf("import-image-%s", image), command); error != nil {
		return fmt.Errorf("Failed to import image %s (%s)", image, error.Error())
	}

	return nil
}

func (deployment *NodeDeployment) Execute(name, command string) (string, error) {
	log.WithFields(log.Fields{"name": name, "node": deployment.name, "_target": deployment.node.IP, "_command": command}).Info("Executing remote command")

	session, error := deployment.getSession()
	if error != nil {
		return "", error
	}

	defer session.Close()

	var buffer bytes.Buffer

	session.Stdout = &buffer

	error = session.Run(command)

	if error != nil {
		error = errors.Wrapf(error, "Could not execute remote command '%s' on '%s'", command, deployment.name)
	}

	return buffer.String(), error
}

func (deployment *NodeDeployment) execute(command string) error {
	session, error := deployment.getSession()
	if error != nil {
		return error
	}

	defer session.Close()

	return session.Run(command)
}

func (deployment *NodeDeployment) UploadFile(from, to string) error {
	deployment.sshLimiter.Lock()
	defer deployment.sshLimiter.Unlock()

	filename := path.Base(to)

	if !utils.FileExists(from) {
		log.WithFields(log.Fields{"name": filename, "node": deployment.name, "_target": deployment.node.IP, "_source-filename": from, "_destination-filename": to}).Debug("Skipping")

		return nil
	}

	executable := false

	info, error := os.Stat(from)
	if error == nil {
		executable = strings.Contains(info.Mode().String(), "x")
	}

	if executable {
		command := fmt.Sprintf("rm %s", to)

		_ = deployment.execute(command)
	}

	log.WithFields(log.Fields{"name": filename, "node": deployment.name, "_target": deployment.node.IP, "_source-filename": from, "_destination-filename": to, "_executable": executable}).Info("Deploying")

	session, error := deployment.getSession()
	if error != nil {
		return error
	}

	defer session.Close()

	if error := scp.CopyPath(from, to, session); error != nil {
		return fmt.Errorf("Could not deploy file '%s' (%s)", from, error.Error())
	}

	return nil
}

func (deployment *NodeDeployment) configureTaint() error {
	kubernetesClient := k8s.NewK8S(deployment.config)

	return kubernetesClient.TaintNode(deployment.name, deployment.node)
}
