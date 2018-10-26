package deployment

import (
	"time"

	"github.com/darxkies/k8s-tew/config"
	"github.com/darxkies/k8s-tew/utils"

	log "github.com/sirupsen/logrus"
)

type Image struct {
	Name     string
	Features config.Features
}

type Deployment struct {
	config            *config.InternalConfig
	identityFile      string
	skipSetup         bool
	skipSetupFeatures config.Features
	pullImages        bool
	forceUpload       bool
	commandRetries    uint
	nodes             map[string]*NodeDeployment
	images            []Image
	parallel          bool
}

func NewDeployment(_config *config.InternalConfig, identityFile string, pullImages bool, forceUpload bool, parallel bool, commandRetries uint, skipSetup, skipStorageSetup, skipMonitoringSetup, skipLoggingSetup, skipBackupSetup, skipShowcaseSetup, skipIngressSetup, skipPackagingSetup bool) *Deployment {
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

	deployment := &Deployment{config: _config, identityFile: identityFile, pullImages: pullImages, forceUpload: forceUpload, parallel: parallel, commandRetries: commandRetries, nodes: nodes, skipSetup: skipSetup, skipSetupFeatures: skipSetupFeatures}

	deployment.images = []Image{
		{Name: deployment.config.Config.Versions.Pause, Features: config.Features{}},
		{Name: deployment.config.Config.Versions.CalicoCNI, Features: config.Features{}},
		{Name: deployment.config.Config.Versions.CalicoNode, Features: config.Features{}},
		{Name: deployment.config.Config.Versions.CalicoTypha, Features: config.Features{}},
		{Name: deployment.config.Config.Versions.CoreDNS, Features: config.Features{}},
		{Name: deployment.config.Config.Versions.MinioServer, Features: config.Features{utils.FeatureBackup, utils.FeatureStorage}},
		{Name: deployment.config.Config.Versions.MinioClient, Features: config.Features{utils.FeatureBackup, utils.FeatureStorage}},
		{Name: deployment.config.Config.Versions.Ark, Features: config.Features{utils.FeatureBackup, utils.FeatureStorage}},
		{Name: deployment.config.Config.Versions.Ceph, Features: config.Features{utils.FeatureStorage}},
		{Name: deployment.config.Config.Versions.CSIAttacher, Features: config.Features{utils.FeatureStorage}},
		{Name: deployment.config.Config.Versions.CSIProvisioner, Features: config.Features{utils.FeatureStorage}},
		{Name: deployment.config.Config.Versions.CSIDriverRegistrar, Features: config.Features{utils.FeatureStorage}},
		{Name: deployment.config.Config.Versions.CSICephRBDPlugin, Features: config.Features{utils.FeatureStorage}},
		{Name: deployment.config.Config.Versions.CSICephFSPlugin, Features: config.Features{utils.FeatureStorage}},
		{Name: deployment.config.Config.Versions.FluentBit, Features: config.Features{utils.FeatureLogging, utils.FeatureStorage}},
		{Name: deployment.config.Config.Versions.Elasticsearch, Features: config.Features{utils.FeatureLogging, utils.FeatureStorage}},
		{Name: deployment.config.Config.Versions.ElasticsearchCron, Features: config.Features{utils.FeatureLogging, utils.FeatureStorage}},
		{Name: deployment.config.Config.Versions.ElasticsearchOperator, Features: config.Features{utils.FeatureLogging, utils.FeatureStorage}},
		{Name: deployment.config.Config.Versions.Kibana, Features: config.Features{utils.FeatureLogging, utils.FeatureStorage}},
		{Name: deployment.config.Config.Versions.Cerebro, Features: config.Features{utils.FeatureLogging, utils.FeatureStorage}},
		{Name: deployment.config.Config.Versions.Heapster, Features: config.Features{utils.FeatureMonitoring, utils.FeatureStorage}},
		{Name: deployment.config.Config.Versions.AddonResizer, Features: config.Features{utils.FeatureMonitoring, utils.FeatureStorage}},
		{Name: deployment.config.Config.Versions.MetricsServer, Features: config.Features{utils.FeatureMonitoring, utils.FeatureStorage}},
		{Name: deployment.config.Config.Versions.KubernetesDashboard, Features: config.Features{}},
		{Name: deployment.config.Config.Versions.PrometheusOperator, Features: config.Features{utils.FeatureMonitoring, utils.FeatureStorage}},
		{Name: deployment.config.Config.Versions.PrometheusConfigReloader, Features: config.Features{utils.FeatureMonitoring, utils.FeatureStorage}},
		{Name: deployment.config.Config.Versions.ConfigMapReload, Features: config.Features{utils.FeatureMonitoring, utils.FeatureStorage}},
		{Name: deployment.config.Config.Versions.KubeStateMetrics, Features: config.Features{utils.FeatureMonitoring, utils.FeatureStorage}},
		{Name: deployment.config.Config.Versions.Grafana, Features: config.Features{utils.FeatureMonitoring, utils.FeatureStorage}},
		{Name: deployment.config.Config.Versions.GrafanaWatcher, Features: config.Features{utils.FeatureMonitoring, utils.FeatureStorage}},
		{Name: deployment.config.Config.Versions.Prometheus, Features: config.Features{utils.FeatureMonitoring, utils.FeatureStorage}},
		{Name: deployment.config.Config.Versions.PrometheusNodeExporter, Features: config.Features{utils.FeatureMonitoring, utils.FeatureStorage}},
		{Name: deployment.config.Config.Versions.PrometheusAlertManager, Features: config.Features{utils.FeatureMonitoring, utils.FeatureStorage}},
		{Name: deployment.config.Config.Versions.CertManagerController, Features: config.Features{utils.FeatureIngress, utils.FeatureStorage}},
		{Name: deployment.config.Config.Versions.NginxIngressDefaultBackend, Features: config.Features{utils.FeatureIngress, utils.FeatureStorage}},
		{Name: deployment.config.Config.Versions.NginxIngressController, Features: config.Features{utils.FeatureIngress, utils.FeatureStorage}},
		{Name: deployment.config.Config.Versions.MySQL, Features: config.Features{utils.FeatureShowcase, utils.FeatureStorage}},
		{Name: deployment.config.Config.Versions.WordPress, Features: config.Features{utils.FeatureShowcase, utils.FeatureStorage}},
	}

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

		if deployment.pullImages {
			// Taint commands
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

func (deployment *Deployment) runPullImages() error {
	if !deployment.pullImages {
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

				return nodeDeployment.pullImage(image.Name)
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

	if error := deployment.runConfigureTaints(); error != nil {
		return error
	}

	if error := deployment.runPullImages(); error != nil {
		return error
	}

	return deployment.runBoostrapperCommands()
}
