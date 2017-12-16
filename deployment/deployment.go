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

	"golang.org/x/crypto/ssh"

	"github.com/tmc/scp"
)

type Deployment struct {
	address      string
	identityFile string
}

func NewDeployment(address string, identityFile string) *Deployment {
	return &Deployment{address: address, identityFile: identityFile}
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
func (deployment *Deployment) CopyFilesTo(files map[string]string) error {
	// Collect remote directories
	directories := map[string]bool{}

	for _, remoteFile := range files {
		directories[path.Dir(remoteFile)] = true
	}

	// Create remote directories
	createDirectoriesCommand := "mkdir -p"

	for directoryName := range directories {
		createDirectoriesCommand += " " + directoryName
	}

	if _, error := deployment.Execute(createDirectoriesCommand); error != nil {
		return error
	}

	// Calculate checksums of remote files
	checksumCommand := "md5sum"

	for _, remoteFile := range files {
		checksumCommand += " " + remoteFile
	}

	output, _ := deployment.Execute(checksumCommand)

	// Parse remote checksum values
	targetFileChecksums := map[string]string{}
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		tokens := strings.Split(line, " ")

		targetFileChecksums[tokens[len(tokens)-1]] = tokens[0]
	}

	// Stop service
	serviceCommand := fmt.Sprintf("systemctl stop %s", utils.SERVICE_NAME)

	_, _ = deployment.Execute(serviceCommand)

	// Copy changed files
	for fromFile, toFile := range files {
		if remoteChecksum, ok := targetFileChecksums[toFile]; ok {
			localChecksum, error := deployment.md5sum(fromFile)

			if error == nil && localChecksum == remoteChecksum {
				continue
			}
		}

		if error := deployment.CopyTo(fromFile, toFile); error != nil {
			return error
		}
	}

	// Registrate and start service
	serviceCommand = fmt.Sprintf("systemctl daemon-reload && systemctl enable %s && systemctl start %s", utils.SERVICE_NAME, utils.SERVICE_NAME)

	_, error := deployment.Execute(serviceCommand)

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

	client, error := ssh.Dial("tcp", deployment.address+":22", &ssh.ClientConfig{
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
	log.WithFields(log.Fields{"command": command, "target": deployment.address}).Info("executing")

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

func (deployment *Deployment) CopyTo(from, to string) error {
	log.WithFields(log.Fields{"source-filename": from, "destination-filename": to, "target": deployment.address}).Info("deploying")

	session, error := deployment.getSession()
	if error != nil {
		return error
	}

	defer session.Close()

	return scp.CopyPath(from, to, session)
}

func Deploy(_config *config.InternalConfig, identityFile string) error {
	for nodeName, node := range _config.Config.Nodes {
		_config.SetNode(nodeName, node)

		deployment := NewDeployment(node.IP, identityFile)

		files := map[string]string{}

		for name, deploymentFile := range _config.Config.DeploymentFiles {
			if !config.CompareLabels(node.Labels, deploymentFile.Labels) {
				continue
			}

			fromFile := _config.GetFullDeploymentFilename(name)
			toFile := _config.GetFullTargetDeploymentFilename(name)

			files[fromFile] = toFile
		}

		if error := deployment.CopyFilesTo(files); error != nil {
			return error
		}
	}

	return nil
}
