package deployment

import (
	"time"

	"github.com/darxkies/k8s-tew/config"
	"github.com/darxkies/k8s-tew/utils"

	log "github.com/sirupsen/logrus"
)

const CONTROLLER_ONLY_TAINT_KEY = "node-role.kubernetes.io/master"

type Deployment struct {
	config         *config.InternalConfig
	identityFile   string
	skipSetup      bool
	skipImagesPull bool
	forceUpload    bool
	commandRetries uint
	nodes          map[string]*NodeDeployment
	images         []string
}

func NewDeployment(_config *config.InternalConfig, identityFile string, skipSetup bool, skipImagesPull bool, forceUpload bool, commandRetries uint) *Deployment {
	nodes := map[string]*NodeDeployment{}

	for nodeName, node := range _config.Config.Nodes {
		nodes[nodeName] = NewNodeDeployment(identityFile, nodeName, node, _config)
	}

	deployment := &Deployment{config: _config, identityFile: identityFile, skipSetup: skipSetup, skipImagesPull: skipImagesPull, forceUpload: forceUpload, commandRetries: commandRetries, nodes: nodes}

	deployment.images = []string{
		utils.GetFullImageName(utils.IMAGE_PAUSE, deployment.config.Config.Versions.Pause),
		utils.GetFullImageName(utils.IMAGE_CALICO_CNI, deployment.config.Config.Versions.CalicoCNI),
		utils.GetFullImageName(utils.IMAGE_CALICO_NODE, deployment.config.Config.Versions.CalicoNode),
		utils.GetFullImageName(utils.IMAGE_CALICO_TYPHA, deployment.config.Config.Versions.CalicoTypha),
		utils.GetFullImageName(utils.IMAGE_COREDNS, deployment.config.Config.Versions.CoreDNS),
		utils.GetFullImageName(utils.IMAGE_MINIO_SERVER, deployment.config.Config.Versions.MinioServer),
		utils.GetFullImageName(utils.IMAGE_MINIO_CLIENT, deployment.config.Config.Versions.MinioClient),
		utils.GetFullImageName(utils.IMAGE_ARK, deployment.config.Config.Versions.Ark),
		utils.GetFullImageName(utils.IMAGE_CEPH, deployment.config.Config.Versions.Ceph),
		utils.GetFullImageName(utils.IMAGE_FLUENT_BIT, deployment.config.Config.Versions.FluentBit),
		utils.GetFullImageName(utils.IMAGE_RBD_PROVISIONER, deployment.config.Config.Versions.RBDProvisioner),
	}

	return deployment
}

func (deployment *Deployment) Steps() int {
	result := 0

	// Create Directories
	result += len(deployment.config.Config.Nodes)

	// Upload Files
	result += len(deployment.config.Config.Nodes)

	// Run Commands
	result += len(deployment.config.Config.Nodes) * len(deployment.config.Config.Commands)

	// Taint commands
	result += len(deployment.config.Config.Nodes)

	return result
}

// Deploy all files to the nodes over SSH
func (deployment *Deployment) Deploy() error {
	sortedNodeKeys := deployment.config.GetSortedNodeKeys()

	for _, nodeName := range sortedNodeKeys {
		nodeDeployment := deployment.nodes[nodeName]

		deployment.config.SetNode(nodeName, nodeDeployment.node)

		if error := nodeDeployment.CreateDirectories(); error != nil {
			return error
		}

		utils.IncreaseProgressStep()

		if error := nodeDeployment.UploadFiles(deployment.forceUpload); error != nil {
			return error
		}

		utils.IncreaseProgressStep()
	}

	return deployment.setup()
}

func (deployment *Deployment) runCommand(name, command string) error {
	var error error

	log.WithFields(log.Fields{"name": name, "_command": command}).Info("Executing command")

	for retries := uint(0); retries < deployment.commandRetries; retries++ {
		// Run command
		if error = utils.RunCommand(command); error == nil {
			break
		}

		log.WithFields(log.Fields{"name": name, "command": command, "error": error}).Debug("Command failed")

		time.Sleep(time.Second)
	}

	if error != nil {
		log.WithFields(log.Fields{"name": name, "command": command, "error": error}).Error("Command failed")

		return error
	}

	return nil
}

func (deployment *Deployment) configureTaint() error {
	var _error error

	sortedNodeKeys := deployment.config.GetSortedNodeKeys()

	for _, nodeName := range sortedNodeKeys {
		nodeDeployment := deployment.nodes[nodeName]

		deployment.config.SetNode(nodeName, nodeDeployment.node)

		log.WithFields(log.Fields{"node": nodeName}).Info("Configuring taint")

		for retries := uint(0); retries < deployment.commandRetries; retries++ {
			if _error = nodeDeployment.configureTaint(); _error == nil {
				break
			}

			time.Sleep(time.Second)
		}

		utils.IncreaseProgressStep()

		if _error != nil {
			log.WithFields(log.Fields{"node": nodeName, "error": _error}).Error("Taint node failed")

			return _error
		}

	}

	if deployment.skipImagesPull {
		return nil
	}

	for _, nodeName := range sortedNodeKeys {
		nodeDeployment := deployment.nodes[nodeName]

		deployment.config.SetNode(nodeName, nodeDeployment.node)

		tasks := utils.Tasks{}

		for _, image := range deployment.images {
			image := image

			tasks = append(tasks, func() error {
				return nodeDeployment.pullImage(image)
			})
		}

		errors := utils.RunParallelTasks(tasks)
		if len(errors) > 0 {
			return errors[0]
		}
	}

	return nil
}

// Run bootstrapper commands
func (deployment *Deployment) setup() error {
	if deployment.skipSetup {
		return nil
	}

	if error := deployment.configureTaint(); error != nil {
		return error
	}

	for _, command := range deployment.config.Config.Commands {
		if !command.Labels.HasLabels([]string{utils.NODE_BOOTSTRAPPER}) {
			utils.IncreaseProgressStep()

			continue
		}

		newCommand, error := deployment.config.ApplyTemplate(command.Name, command.Command)
		if error != nil {
			return error
		}

		if error := deployment.runCommand(command.Name, newCommand); error != nil {
			return error
		}

		utils.IncreaseProgressStep()
	}

	return nil
}
