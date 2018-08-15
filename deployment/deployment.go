package deployment

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/darxkies/k8s-tew/config"
	"github.com/darxkies/k8s-tew/utils"

	log "github.com/sirupsen/logrus"

	"golang.org/x/crypto/ssh"

	"github.com/tmc/scp"
)

type Deployment struct {
	identityFile string
	node         *config.Node
	config       *config.InternalConfig
}

func NewDeployment(identityFile string, node *config.Node, config *config.InternalConfig) *Deployment {
	return &Deployment{identityFile: identityFile, node: node, config: config}
}

func (deployment *Deployment) md5sum(filename string) (result string, error error) {
	file, error := os.Open(filename)

	if error != nil {
		return
	}

	defer file.Close()

	hash := md5.New()

	if _, error = io.Copy(hash, file); error != nil {
		return
	}

	result = hex.EncodeToString(hash.Sum(nil)[:16])

	return
}

func (deployment *Deployment) CreateDirectories() error {
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

	// Create remote directories
	createDirectoriesCommand := "mkdir -p"

	for directoryName := range directories {
		createDirectoriesCommand += " " + directoryName
	}

	if _, error := deployment.Execute(createDirectoriesCommand); error != nil {
		return error
	}

	return nil
}

func (deployment *Deployment) getFiles() map[string]string {
	files := map[string]string{}

	// Collect files to be deployed
	for name, file := range deployment.config.Config.Assets.Files {
		if !config.CompareLabels(deployment.node.Labels, file.Labels) {
			continue
		}

		fromFile := deployment.config.GetFullLocalAssetFilename(name)
		toFile := deployment.config.GetFullTargetAssetFilename(name)

		files[fromFile] = toFile
	}

	return files
}

func (deployment *Deployment) getRemoteFileChecksums() map[string]string {
	// Calculate checksums of remote files
	checksumCommand := "md5sum"

	for _, toFile := range deployment.getFiles() {
		checksumCommand += " " + toFile
	}

	output, _ := deployment.Execute(checksumCommand)

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

func (deployment *Deployment) getChangedFiles() map[string]string {
	remoteFileChecksums := deployment.getRemoteFileChecksums()

	files := map[string]string{}

	for fromFile, toFile := range deployment.getFiles() {
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

func (deployment *Deployment) UploadFiles() error {
	changedFiles := deployment.getChangedFiles()

	if len(changedFiles) == 0 {
		return nil
	}

	// Stop service
	_, _ = deployment.Execute(fmt.Sprintf("systemctl stop %s", utils.SERVICE_NAME))

	// Copy changed files
	for fromFile, toFile := range changedFiles {
		if error := deployment.UploadFile(fromFile, toFile); error != nil {
			return error
		}
	}

	// Registrate and start service
	_, error := deployment.Execute(fmt.Sprintf("systemctl daemon-reload && systemctl enable %s && systemctl start %s", utils.SERVICE_NAME, utils.SERVICE_NAME))

	return error
}

func (deployment *Deployment) getSession() (*ssh.Session, error) {
	privateKeyContent, error := ioutil.ReadFile(deployment.identityFile)
	if error != nil {
		return nil, error
	}

	privateKey, error := ssh.ParsePrivateKey(privateKeyContent)
	if error != nil {
		return nil, error
	}

	client, error := ssh.Dial("tcp", deployment.node.IP+":22", &ssh.ClientConfig{
		User: utils.DEPLOYMENT_USER,
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

func (deployment *Deployment) Execute(command string) (string, error) {
	log.WithFields(log.Fields{"target": deployment.node.IP, "command": command}).Info("executing remote command")

	session, error := deployment.getSession()
	if error != nil {
		return "", error
	}

	defer session.Close()

	var buffer bytes.Buffer

	session.Stdout = &buffer

	error = session.Run(command)

	return buffer.String(), error
}

func (deployment *Deployment) UploadFile(from, to string) error {
	log.WithFields(log.Fields{"target": deployment.node.IP, "source-filename": from, "destination-filename": to}).Info("deploying")

	session, error := deployment.getSession()
	if error != nil {
		return error
	}

	defer session.Close()

	return scp.CopyPath(from, to, session)
}

func Steps(_config *config.InternalConfig) int {
	result := 0

	// Create Directories
	result += len(_config.Config.Nodes)

	// Upload Files
	result += len(_config.Config.Nodes)

	// Run Commands
	result += len(_config.Config.Nodes) * len(_config.Config.Commands)

	return result
}

// Deploy all files to the nodes over SSH
func Deploy(_config *config.InternalConfig, identityFile string) error {
	sortedNodeKeys := _config.GetSortedNodeKeys()

	for _, nodeName := range sortedNodeKeys {
		node := _config.Config.Nodes[nodeName]

		_config.SetNode(nodeName, node)

		deployment := NewDeployment(identityFile, node, _config)

		if error := deployment.CreateDirectories(); error != nil {
			return error
		}

		utils.IncreaseProgressStep()

		if error := deployment.UploadFiles(); error != nil {
			return error
		}

		utils.IncreaseProgressStep()
	}

	return nil
}

// Run bootstrapper commands
func Setup(_config *config.InternalConfig, commandRetries uint) error {
	for _, command := range _config.Config.Commands {
		if !command.Labels.HasLabels([]string{utils.NODE_BOOTSTRAPPER}) {
			utils.IncreaseProgressStep()

			continue
		}

		newCommand, error := _config.ApplyTemplate(command.Name, command.Command)
		if error != nil {
			return error
		}

		log.WithFields(log.Fields{"name": command.Name, "command": newCommand}).Info("executing command")

		for retries := uint(0); retries < commandRetries; retries++ {
			// Run command
			error = utils.RunCommand(newCommand)
			if error == nil {
				break
			}

			time.Sleep(time.Second)
		}

		if error != nil {
			log.WithFields(log.Fields{"name": command.Name, "command": newCommand, "error": error}).Error("command failed")

			return error
		}

		utils.IncreaseProgressStep()
	}

	return nil
}
