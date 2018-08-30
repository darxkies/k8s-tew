package servers

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/darxkies/k8s-tew/utils"

	"github.com/darxkies/k8s-tew/config"

	log "github.com/sirupsen/logrus"
)

type Servers struct {
	config      *config.InternalConfig
	servers     []Server
	stop        bool
	killTimeout uint
}

func NewServers(_config *config.InternalConfig, killTimeout uint) *Servers {
	return &Servers{config: _config, servers: []Server{}, stop: false, killTimeout: killTimeout}
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

func (servers *Servers) addVIPManager(enabled bool, virtualIP, virtualIPInterface, nodeName, nodeIP, nodeRole string, raftPort uint16) {
	if !enabled {
		return
	}

	if len(virtualIP) == 0 {
		return
	}

	if len(virtualIPInterface) == 0 {
		return
	}

	peers := Peers{}

	for nodeName, node := range servers.config.Config.Nodes {
		if !node.Labels.HasLabels(config.Labels{nodeRole}) {
			continue
		}

		peers[nodeName] = fmt.Sprintf("%s:%d", node.IP, raftPort)
	}

	logger := Logger{}

	servers.add(NewVIPManager(nodeRole, nodeName, fmt.Sprintf("%s:%d", nodeIP, raftPort), virtualIP, peers, logger, virtualIPInterface))
}

func (servers *Servers) Run(commandRetries uint) error {
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

	// Add Controllers/Workers VIP servers
	servers.addVIPManager(servers.config.Node.IsController(), servers.config.Config.ControllerVirtualIP, servers.config.Config.ControllerVirtualIPInterface, servers.config.Name, servers.config.Node.IP, utils.NODE_CONTROLLER, servers.config.Config.VIPRaftControllerPort)
	servers.addVIPManager(servers.config.Node.IsWorker(), servers.config.Config.WorkerVirtualIP, servers.config.Config.WorkerVirtualIPInterface, servers.config.Name, servers.config.Node.IP, utils.NODE_WORKER, servers.config.Config.VIPRaftWorkerPort)

	// Start servers
	for _, server := range servers.servers {
		if error := server.Start(); error != nil {
			log.WithFields(log.Fields{"name": server.Name(), "error": error}).Error("Server start failed")

			return error
		}

		utils.IncreaseProgressStep()
	}

	// Register servers' stop
	defer func() {
		for _, server := range servers.servers {
			log.WithFields(log.Fields{"name": server.Name()}).Info("Stopping server")

			server.Stop()
		}

		log.Info("Stopped all servers")

		utils.KillProcessChildren(os.Getpid(), servers.killTimeout)
	}()

	go func() {
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

			if error := servers.runCommand(command, commandRetries, index+1, len(servers.config.Config.Commands)); error != nil {
				log.WithFields(log.Fields{"error": error}).Fatal("Cluster setup failed")
			}

			utils.IncreaseProgressStep()
		}

		log.Info("Cluster setup finished")

		utils.HideProgress()
	}()

	// Wait for signals to stop
	signals := make(chan os.Signal, 1)

	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	<-signals

	servers.stop = true

	return nil
}
