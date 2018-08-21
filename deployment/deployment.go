package deployment

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"github.com/darxkies/k8s-tew/config"
	"github.com/darxkies/k8s-tew/utils"

	log "github.com/sirupsen/logrus"

	"golang.org/x/crypto/ssh"

	"github.com/tmc/scp"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
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

	if _, error := deployment.Execute("create-directories", createDirectoriesCommand); error != nil {
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
	_, _ = deployment.Execute("stop-service", fmt.Sprintf("systemctl stop %s", utils.SERVICE_NAME))

	// Copy changed files
	for fromFile, toFile := range changedFiles {
		if error := deployment.UploadFile(fromFile, toFile); error != nil {
			return error
		}
	}

	// Registrate and start service
	_, error := deployment.Execute("start-service", fmt.Sprintf("systemctl daemon-reload && systemctl enable %s && systemctl start %s", utils.SERVICE_NAME, utils.SERVICE_NAME))

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

func (deployment *Deployment) Execute(name, command string) (string, error) {
	log.WithFields(log.Fields{"name": name, "target": deployment.node.IP, "_command": command}).Info("Executing remote command")

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
	filename := path.Base(to)

	log.WithFields(log.Fields{"name": filename, "target": deployment.node.IP, "_source-filename": from, "_destination-filename": to}).Info("Deploying")

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

	// Taint commands
	result += len(_config.Config.Nodes)

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

func taintNode(kubeconfig, nodeName string, isControllerOnly bool) error {
	// Configure connection
	config, error := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if error != nil {
		return error
	}

	// Create client
	clientset, error := kubernetes.NewForConfig(config)
	if error != nil {
		return error
	}

	// Get Node
	node, error := clientset.CoreV1().Nodes().Get(nodeName, metav1.GetOptions{})
	if error != nil {
		return error
	}

	changed := false

	if isControllerOnly {
		found := false

		for _, taint := range node.Spec.Taints {
			if taint.Key == CONTROLLER_ONLY_TAINT_KEY {
				found = true

				break
			}
		}

		if !found {
			node.Spec.Taints = append(node.Spec.Taints, v1.Taint{Key: CONTROLLER_ONLY_TAINT_KEY, Value: "true", Effect: v1.TaintEffectNoSchedule})

			changed = true
		}
	} else {
		taints := []v1.Taint{}

		for _, taint := range node.Spec.Taints {
			if taint.Key == CONTROLLER_ONLY_TAINT_KEY {
				changed = true

				continue
			}

			taints = append(taints, taint)
		}

		node.Spec.Taints = taints
	}

	if !changed {
		return nil
	}

	_, error = clientset.CoreV1().Nodes().Update(node)

	return error
}

func runCommand(name, command string, commandRetries uint) error {
	var error error

	log.WithFields(log.Fields{"name": name, "_command": command}).Info("Executing command")

	for retries := uint(0); retries < commandRetries; retries++ {
		// Run command
		if error = utils.RunCommand(command); error == nil {
			break
		}

		time.Sleep(time.Second)
	}

	if error != nil {
		log.WithFields(log.Fields{"name": name, "command": command, "error": error}).Error("Command failed")

		return error
	}

	return nil
}

const CONTROLLER_ONLY_TAINT_KEY = "node-role.kubernetes.io/master"

// Run bootstrapper commands
func Setup(_config *config.InternalConfig, commandRetries uint) error {
	kubeconfig := _config.GetFullLocalAssetFilename(utils.ADMIN_KUBECONFIG)

	var error error

	for nodeName, node := range _config.Config.Nodes {
		log.WithFields(log.Fields{"node": nodeName}).Info("Taint node")

		for retries := uint(0); retries < commandRetries; retries++ {
			if error = taintNode(kubeconfig, nodeName, node.IsControllerOnly()); error == nil {
				break
			}

			time.Sleep(time.Second)
		}

		utils.IncreaseProgressStep()

		if error != nil {
			log.WithFields(log.Fields{"node": nodeName, "error": error}).Error("Taint node failed")

			return error
		}
	}

	for _, command := range _config.Config.Commands {
		if !command.Labels.HasLabels([]string{utils.NODE_BOOTSTRAPPER}) {
			utils.IncreaseProgressStep()

			continue
		}

		newCommand, error := _config.ApplyTemplate(command.Name, command.Command)
		if error != nil {
			return error
		}

		if error := runCommand(command.Name, newCommand, commandRetries); error != nil {
			return error
		}

		utils.IncreaseProgressStep()
	}

	return nil
}
