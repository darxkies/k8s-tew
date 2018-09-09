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
	pullImages     bool
	forceUpload    bool
	commandRetries uint
	nodes          map[string]*NodeDeployment
	images         []string
	parallel       bool
}

func NewDeployment(_config *config.InternalConfig, identityFile string, skipSetup bool, pullImages bool, forceUpload bool, parallel bool, commandRetries uint) *Deployment {
	nodes := map[string]*NodeDeployment{}

	for nodeName, node := range _config.Config.Nodes {
		nodes[nodeName] = NewNodeDeployment(identityFile, nodeName, node, _config, parallel)
	}

	deployment := &Deployment{config: _config, identityFile: identityFile, skipSetup: skipSetup, pullImages: pullImages, forceUpload: forceUpload, parallel: parallel, commandRetries: commandRetries, nodes: nodes}

	deployment.images = []string{
		deployment.config.Config.Versions.Pause,
		deployment.config.Config.Versions.CalicoCNI,
		deployment.config.Config.Versions.CalicoNode,
		deployment.config.Config.Versions.CalicoTypha,
		deployment.config.Config.Versions.CoreDNS,
		deployment.config.Config.Versions.MinioServer,
		deployment.config.Config.Versions.MinioClient,
		deployment.config.Config.Versions.Ark,
		deployment.config.Config.Versions.Ceph,
		deployment.config.Config.Versions.FluentBit,
		deployment.config.Config.Versions.RBDProvisioner,
		deployment.config.Config.Versions.Elasticsearch,
		deployment.config.Config.Versions.ElasticsearchCron,
		deployment.config.Config.Versions.ElasticsearchOperator,
		deployment.config.Config.Versions.Kibana,
		deployment.config.Config.Versions.Cerebro,
		deployment.config.Config.Versions.Heapster,
		deployment.config.Config.Versions.AddonResizer,
		deployment.config.Config.Versions.KubernetesDashboard,
		deployment.config.Config.Versions.CertManagerController,
		deployment.config.Config.Versions.NginxIngressDefaultBackend,
		deployment.config.Config.Versions.NginxIngressController,
		deployment.config.Config.Versions.MetricsServer,
		deployment.config.Config.Versions.PrometheusOperator,
		deployment.config.Config.Versions.PrometheusConfigReloader,
		deployment.config.Config.Versions.ConfigMapReload,
		deployment.config.Config.Versions.KubeStateMetrics,
		deployment.config.Config.Versions.Grafana,
		deployment.config.Config.Versions.GrafanaWatcher,
		deployment.config.Config.Versions.Prometheus,
		deployment.config.Config.Versions.PrometheusNodeExporter,
		deployment.config.Config.Versions.PrometheusAlertManager,
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

				return nodeDeployment.pullImage(image)
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
