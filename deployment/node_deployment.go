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

	"github.com/darxkies/k8s-tew/config"
	"github.com/darxkies/k8s-tew/utils"
	log "github.com/sirupsen/logrus"
	"github.com/tmc/scp"
	"golang.org/x/crypto/ssh"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const CONCURRENT_SSH_CONNECTIONS_LIMIT = 10

type NodeDeployment struct {
	identityFile string
	name         string
	node         *config.Node
	config       *config.InternalConfig
	sshLimiter   *utils.Limiter
}

func NewNodeDeployment(identityFile string, name string, node *config.Node, config *config.InternalConfig) *NodeDeployment {
	return &NodeDeployment{identityFile: identityFile, name: name, node: node, config: config, sshLimiter: utils.NewLimiter(CONCURRENT_SSH_CONNECTIONS_LIMIT)}
}

func (deployment *NodeDeployment) md5sum(filename string) (result string, error error) {
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

func (deployment *NodeDeployment) CreateDirectories() error {
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

func (deployment *NodeDeployment) getFiles() map[string]string {
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

func (deployment *NodeDeployment) getRemoteFileChecksums() map[string]string {
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

func (deployment *NodeDeployment) getChangedFiles() map[string]string {
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

func (deployment *NodeDeployment) UploadFiles(forceUpload bool) error {
	var files map[string]string

	if forceUpload {
		files = deployment.getFiles()
	} else {
		files = deployment.getChangedFiles()
	}

	if len(files) == 0 {
		return nil
	}

	// Stop service
	_, _ = deployment.Execute("stop-service", fmt.Sprintf("systemctl stop %s", utils.SERVICE_NAME))

	tasks := utils.Tasks{}

	// Copy changed files
	for fromFile, toFile := range files {
		fromFile := fromFile
		toFile := toFile

		tasks = append(tasks, func() error {
			return deployment.UploadFile(fromFile, toFile)
		})
	}

	// Upload files
	if errors := utils.RunParallelTasks(tasks); len(errors) > 0 {
		return errors[0]
	}

	// Registrate and start service
	_, error := deployment.Execute("start-service", fmt.Sprintf("systemctl daemon-reload && systemctl enable %s && systemctl start %s", utils.SERVICE_NAME, utils.SERVICE_NAME))

	return error
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

func (deployment *NodeDeployment) pullImage(image string) error {
	deployment.sshLimiter.Lock()
	defer deployment.sshLimiter.Unlock()

	crictl := deployment.config.GetFullTargetAssetFilename(utils.CRICTL_BINARY)
	containerdSock := deployment.config.GetFullTargetAssetFilename(utils.CONTAINERD_SOCK)
	command := fmt.Sprintf("CONTAINER_RUNTIME_ENDPOINT=unix://%s %s pull %s", containerdSock, crictl, image)

	output, error := deployment.Execute(fmt.Sprintf("pull-image-%s", image), command)
	if error != nil {
		return fmt.Errorf("%s (%s)", error.Error(), output)
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

	return buffer.String(), error
}

func (deployment *NodeDeployment) UploadFile(from, to string) error {
	deployment.sshLimiter.Lock()
	defer deployment.sshLimiter.Unlock()

	filename := path.Base(to)

	log.WithFields(log.Fields{"name": filename, "node": deployment.name, "_target": deployment.node.IP, "_source-filename": from, "_destination-filename": to}).Info("Deploying")

	session, error := deployment.getSession()
	if error != nil {
		return error
	}

	defer session.Close()

	return scp.CopyPath(from, to, session)
}

func (deployment *NodeDeployment) configureTaint() error {
	kubeconfig := deployment.config.GetFullLocalAssetFilename(utils.ADMIN_KUBECONFIG)

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
	node, error := clientset.CoreV1().Nodes().Get(deployment.name, metav1.GetOptions{})
	if error != nil {
		return error
	}

	changed := false

	if deployment.node.IsControllerOnly() {
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
