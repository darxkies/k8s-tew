package servers

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/darxkies/k8s-tew/pkg/config"
	"github.com/darxkies/k8s-tew/pkg/container"
	"github.com/darxkies/k8s-tew/pkg/deployment"
	"github.com/darxkies/k8s-tew/pkg/k8s"
	"github.com/darxkies/k8s-tew/pkg/utils"
)

type Servers struct {
	config  *config.InternalConfig
	servers []Server
	stop    bool
}

func NewServers(_config *config.InternalConfig) *Servers {
	return &Servers{config: _config, servers: []Server{}, stop: false}
}

func (servers *Servers) add(server Server) {
	servers.servers = append(servers.servers, server)
}

func (servers *Servers) runCommand(command *config.Command, commandRetries uint, step, count int) error {
	newCommand, error := servers.config.ApplyTemplate(command.Name, command.Command)
	if error != nil {
		return error
	}

	log.WithFields(log.Fields{"name": command.Name, "_command": newCommand}).Info("Executing command")

	for retries := uint(0); retries < commandRetries; retries++ {
		if servers.stop {
			break
		}

		// Run command
		if error = utils.RunCommand(newCommand); error == nil {
			break
		}

		time.Sleep(time.Second)
	}

	if error != nil {
		log.WithFields(log.Fields{"name": command.Name, "command": newCommand, "error": error}).Error("Command failed")

		return error
	}

	return nil
}

func (servers *Servers) Steps() int {
	return len(servers.config.Config.Servers) + len(servers.config.Config.Commands) + 1
}

func (servers *Servers) Run(commandRetries uint, cleanup func()) error {
	isContainerd := func(server Server) bool {
		return server.Name() == utils.ContainerdServerName
	}

	pathEnvironment := os.Getenv("PATH")
	pathEnvironment = fmt.Sprintf("PATH=%s:%s", servers.config.GetFullLocalAssetDirectory(utils.DirectoryHostBinaries), pathEnvironment)

	// Add servers
	for _, serverConfig := range servers.config.Config.Servers {
		if !serverConfig.Enabled {
			continue
		}

		if !config.CompareLabels(servers.config.Node.Labels, serverConfig.Labels) {
			continue
		}

		server, error := NewServerWrapper(*servers.config, serverConfig.Name, serverConfig, pathEnvironment)

		if error != nil {
			return errors.Wrapf(error, "server wrapper for '%s' failed", serverConfig.Name)
		}

		servers.add(server)
	}

	// Start only containerd
	for _, server := range servers.servers {
		if !isContainerd(server) {
			continue
		}

		if error := server.Start(); error != nil {
			return error
		}

		utils.IncreaseProgressStep()
	}

	// Restore Pods
	_pods := container.NewPods(servers.config)

	_pods.Restore()

	// Start all other entries
	for _, server := range servers.servers {
		if isContainerd(server) {
			continue
		}

		if error := server.Start(); error != nil {
			return error
		}

		utils.IncreaseProgressStep()
	}

	kubernetesClient := k8s.NewK8S(servers.config)

	go func() {
		log.Info("Uncordoning")

		for {
			if _error := kubernetesClient.Uncordon(servers.config.Name); _error != nil {
				log.WithFields(log.Fields{"status": _error}).Debug("Uncordoning")

				time.Sleep(time.Second)

				continue
			}

			break
		}

		log.Info("Uncordoned")
	}()

	// Register servers' stop
	defer func() {
		log.Info("Cordoning")

		if _error := kubernetesClient.Cordon(servers.config.Name); _error != nil {
			log.WithFields(log.Fields{"Error": _error}).Error("Cordoning failed")

		} else {
			log.Info("Cordoned")

			log.Info("Draining")

			if _error := kubernetesClient.Drain(servers.config.Name); _error != nil {
				log.WithFields(log.Fields{"error": _error}).Error("Drain failed")

			} else {
				log.Info("Drained")
			}
		}

		// Stop all servers but containerd
		for _, server := range servers.servers {
			if isContainerd(server) {
				continue
			}

			server.Stop()
		}

		cleanup()

		// Stop containerd
		for _, server := range servers.servers {
			server.Stop()
		}

		log.Info("Stopped all servers")
	}()

	// Import images if downloaded and if node is a Bootstrapper
	if config.CompareLabels(servers.config.Node.Labels, config.Labels{utils.NodeBootstrapper}) {
		go func() {
			for _, image := range servers.config.Config.Versions.GetImages() {
				command := deployment.GetImportImageCommand(servers.config, image.Name, servers.config.GetFullTargetAssetFilename(image.GetImageFilename()))

				log.WithFields(log.Fields{"name": image.Name}).Info("Import image")

				for {
					if _error := utils.RunCommand(command); _error != nil {
						log.WithFields(log.Fields{"error": _error}).Info("Image import failed")

						time.Sleep(time.Second)

						continue
					}

					break
				}
			}
		}()
	}

	go func() {
		successful := true

		// Register commands based on labels to be executed asynchronously
		for index, command := range servers.config.Config.Commands {
			if !config.CompareLabels(servers.config.Node.Labels, command.Labels) {
				utils.IncreaseProgressStep()

				continue
			}

			if !utils.HasOS(command.OS) {
				utils.IncreaseProgressStep()

				continue
			}

			if len(command.Manifest) > 0 {
				if error := k8s.ApplyManifest(servers.config, command.Name, command.Manifest, -1); error != nil {
					log.WithFields(log.Fields{"error": error}).Error("Cluster setup failed")

					successful = false

					servers.stop = true

					break
				}

			} else {
				if error := servers.runCommand(command, commandRetries, index+1, len(servers.config.Config.Commands)); error != nil {
					log.WithFields(log.Fields{"error": error}).Error("Cluster setup failed")

					successful = false

					servers.stop = true

					break
				}
			}

			utils.IncreaseProgressStep()
		}

		if successful {
			log.Info("Cluster setup finished - Supervising servers")
		}

		utils.HideProgress()
	}()

	// Wait for signals to stop
	signals := make(chan os.Signal, 1)

	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	<-signals

	servers.stop = true

	return nil
}
