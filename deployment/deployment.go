package deployment

import (
	"time"

	"github.com/darxkies/k8s-tew/config"
	"github.com/darxkies/k8s-tew/utils"

	log "github.com/sirupsen/logrus"
)

type Deployment struct {
	config            *config.InternalConfig
	identityFile      string
	skipSetup         bool
	skipSetupFeatures config.Features
	forceUpload       bool
	commandRetries    uint
	nodes             map[string]*NodeDeployment
	images            config.Images
	parallel          bool
	importImages      bool
}

func NewDeployment(_config *config.InternalConfig, identityFile string, importImages, forceUpload bool, parallel bool, commandRetries uint, skipSetup, skipStorageSetup, skipMonitoringSetup, skipLoggingSetup, skipBackupSetup, skipShowcaseSetup, skipIngressSetup, skipPackagingSetup bool) *Deployment {
	nodes := map[string]*NodeDeployment{}

	for nodeName, node := range _config.Config.Nodes {
		nodes[nodeName] = NewNodeDeployment(identityFile, nodeName, node, _config, parallel)
	}

	skipSetupFeatures := config.Features{}

	if skipStorageSetup {
		skipSetupFeatures = append(skipSetupFeatures, utils.FeatureStorage)
	}

	if skipMonitoringSetup {
		skipSetupFeatures = append(skipSetupFeatures, utils.FeatureMonitoring)
	}

	if skipLoggingSetup {
		skipSetupFeatures = append(skipSetupFeatures, utils.FeatureLogging)
	}

	if skipBackupSetup {
		skipSetupFeatures = append(skipSetupFeatures, utils.FeatureBackup)
	}

	if skipShowcaseSetup {
		skipSetupFeatures = append(skipSetupFeatures, utils.FeatureShowcase)
	}

	if skipIngressSetup {
		skipSetupFeatures = append(skipSetupFeatures, utils.FeatureIngress)
	}

	if skipPackagingSetup {
		skipSetupFeatures = append(skipSetupFeatures, utils.FeaturePackaging)
	}

	deployment := &Deployment{config: _config, identityFile: identityFile, importImages: importImages, forceUpload: forceUpload, parallel: parallel, commandRetries: commandRetries, nodes: nodes, skipSetup: skipSetup, skipSetupFeatures: skipSetupFeatures}

	deployment.images = deployment.config.Config.Versions.GetImages()

	return deployment
}

func (deployment *Deployment) Steps() int {
	result := 0

	// Files deployment
	for _, node := range deployment.nodes {
		result += node.Steps()
	}

	if !deployment.skipSetup {
		// Taint commands
		result += len(deployment.config.Config.Nodes)

		if deployment.importImages {
			// Import images
			result += len(deployment.config.Config.Nodes) * len(deployment.images)
		}

		// Run Commands
		result += len(deployment.config.Config.Nodes) * len(deployment.config.Config.Commands)

	}

	return result
}

// Deploy all files to the nodes over SSH
func (deployment *Deployment) Deploy() error {
	sortedNodeKeys := deployment.config.GetSortedNodeKeys()

	for _, nodeName := range sortedNodeKeys {
		nodeDeployment := deployment.nodes[nodeName]

		deployment.config.SetNode(nodeName, nodeDeployment.node)

		if error := nodeDeployment.UploadFiles(deployment.forceUpload); error != nil {
			return error
		}
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

func (deployment *Deployment) runConfigureTaints() error {
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

	return nil
}

func (deployment *Deployment) runImportImages() error {
	if !deployment.importImages {
		return nil
	}

	sortedNodeKeys := deployment.config.GetSortedNodeKeys()

	for _, nodeName := range sortedNodeKeys {
		nodeDeployment := deployment.nodes[nodeName]

		deployment.config.SetNode(nodeName, nodeDeployment.node)

		tasks := utils.Tasks{}

		for _, image := range deployment.images {
			image := image

			tasks = append(tasks, func() error {
				defer utils.IncreaseProgressStep()

				if image.Features.HasFeatures(deployment.skipSetupFeatures) {
					return nil
				}

				_ = nodeDeployment.importImage(image.Name, deployment.config.GetFullTargetAssetFilename(image.GetImageFilename()))

				return nil
			})
		}

		if errors := utils.RunParallelTasks(tasks, deployment.parallel); len(errors) > 0 {
			return errors[0]
		}
	}

	return nil
}

// Run bootstrapper commands
func (deployment *Deployment) runBoostrapperCommands() error {
	for _, command := range deployment.config.Config.Commands {
		if !command.Labels.HasLabels([]string{utils.NodeBootstrapper}) {
			utils.IncreaseProgressStep()

			continue
		}

		if command.Features.HasFeatures(deployment.skipSetupFeatures) {
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

// Setup nodes
func (deployment *Deployment) setup() error {
	if deployment.skipSetup {
		return nil
	}

	if error := deployment.runImportImages(); error != nil {
		return error
	}

	if error := deployment.runConfigureTaints(); error != nil {
		return error
	}

	return deployment.runBoostrapperCommands()
}
