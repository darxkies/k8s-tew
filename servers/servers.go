package servers

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/darxkies/k8s-tew/utils"

	"github.com/darxkies/k8s-tew/config"

	log "github.com/sirupsen/logrus"
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

	log.WithFields(log.Fields{"name": command.Name, "command": newCommand}).Info("executing command")

	stop := utils.ShowProgress("cluster setup", step, count)
	defer stop()

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
		log.WithFields(log.Fields{"name": command.Name, "command": newCommand, "error": error}).Error("command failed")

		return error
	}

	return nil
}

func (servers *Servers) Run(commandRetries uint) error {
	// Dump configuration
	servers.config.Dump()

	// Add servers
	for _, serverConfig := range servers.config.Config.Servers {
		if !serverConfig.Enabled {
			continue
		}

		if !config.CompareLabels(servers.config.Node.Labels, serverConfig.Labels) {
			continue
		}

		server, error := NewServerWrapper(*servers.config, serverConfig.Name, serverConfig)

		if error != nil {
			return error
		}

		servers.add(server)
	}

	// Add Controller VIP Manager
	if servers.config.Node.IsController() && len(servers.config.Config.ControllerVirtualIP) > 0 && len(servers.config.Config.ControllerVirtualIPInterface) > 0 {
		servers.add(NewVIPManager(utils.ELECTION_CONTROLLER, servers.config.Node.IP, servers.config.Config.ControllerVirtualIP, servers.config.Config.ControllerVirtualIPInterface, servers.config.GetETCDClientEndpoints(), servers.config.GetFullLocalAssetFilename(utils.CA_PEM), servers.config.GetFullLocalAssetFilename(utils.VIRTUAL_IP_PEM), servers.config.GetFullLocalAssetFilename(utils.VIRTUAL_IP_KEY_PEM)))
	}

	// Add Worker VIP Manager
	if servers.config.Node.IsWorker() && len(servers.config.Config.WorkerVirtualIP) > 0 && len(servers.config.Config.WorkerVirtualIPInterface) > 0 {
		servers.add(NewVIPManager(utils.ELECTION_WORKER, servers.config.Node.IP, servers.config.Config.WorkerVirtualIP, servers.config.Config.WorkerVirtualIPInterface, servers.config.GetETCDClientEndpoints(), servers.config.GetFullLocalAssetFilename(utils.CA_PEM), servers.config.GetFullLocalAssetFilename(utils.VIRTUAL_IP_PEM), servers.config.GetFullLocalAssetFilename(utils.VIRTUAL_IP_KEY_PEM)))
	}

	// Start servers
	for _, server := range servers.servers {
		if error := server.Start(); error != nil {
			log.WithFields(log.Fields{"name": server.Name(), "error": error}).Error("server start failed")

			return error
		}

	}

	// Register servers' stop
	defer func() {
		for _, server := range servers.servers {
			log.WithFields(log.Fields{"name": server.Name()}).Info("stopping server")

			server.Stop()
		}

		log.Info("stopped all servers")
	}()

	go func() {
		// Register commands based on labels to be executed asynchronously
		for index, command := range servers.config.Config.Commands {
			if !config.CompareLabels(servers.config.Node.Labels, command.Labels) {
				continue
			}

			if !utils.HasOS(command.OS) {
				continue
			}

			if error := servers.runCommand(command, commandRetries, index+1, len(servers.config.Config.Commands)); error != nil {
				log.WithFields(log.Fields{"error": error}).Fatal("cluster setup failed")
			}
		}

		log.Info("cluster setup finished")
	}()

	// Wait for signals to stop
	signals := make(chan os.Signal, 1)

	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	<-signals

	servers.stop = true

	return nil
}
