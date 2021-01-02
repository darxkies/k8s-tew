package servers

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/darxkies/k8s-tew/pkg/config"
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

func (servers *Servers) extractEmbeddedFiles() error {
	log.Debug("Extracting embedded files")

	return utils.GetEmbeddedFiles(func(filename string, in io.ReadCloser) error {
		log.WithFields(log.Fields{"filename": filename}).Info("Extracting embedded file")

		hostDirectory := servers.config.GetFullLocalAssetDirectory(utils.DirectoryHostBinaries)
		outFilename := path.Join(hostDirectory, filename)

		if error := utils.CreateDirectoryIfMissing(path.Dir(outFilename)); error != nil {
			return error
		}

		// Defer source file closing
		defer in.Close()

		// Open target file
		out, error := os.OpenFile(outFilename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
		if error != nil {
			return error
		}

		// Defer target file closing
		defer out.Close()

		// Copy file content
		if _, error = io.Copy(out, in); error != nil {
			return error
		}

		// Sync content to storage
		return out.Sync()
	})
}

func (servers *Servers) Run(commandRetries uint, cleanup func()) error {
	// Make sure the embedded dependencies are in place before the servers are started
	if error := servers.extractEmbeddedFiles(); error != nil {
		return errors.Wrap(error, "extracting embedded files failed")
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

	// Start servers
	for _, server := range servers.servers {
		if error := server.Start(); error != nil {
			log.WithFields(log.Fields{"name": server.Name(), "error": error}).Error("Server start failed")

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

		for _, server := range servers.servers {
			if server.Name() == utils.ContainerdServerName {
				continue
			}

			log.WithFields(log.Fields{"name": server.Name()}).Info("Stopping server")

			server.Stop()
		}

		cleanup()

		for _, server := range servers.servers {
			if server.Name() != utils.ContainerdServerName {
				continue
			}

			log.WithFields(log.Fields{"name": server.Name()}).Info("Stopping server")

			server.Stop()
		}

		log.Info("Stopped all servers")
	}()

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
				if error := k8s.ApplyManifest(servers.config, command.Name, command.Manifest, commandRetries); error != nil {
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
