package servers

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/darxkies/k8s-tew/utils"

	"github.com/darxkies/k8s-tew/config"

	log "github.com/sirupsen/logrus"
)

type Server interface {
	Start() error
	Stop()
	Name() string
}

type Servers struct {
	config  *config.InternalConfig
	servers []Server
}

func NewServers(_config *config.InternalConfig) *Servers {
	return &Servers{config: _config, servers: []Server{}}
}

func (servers *Servers) add(server Server) {
	servers.servers = append(servers.servers, server)
}

func (servers *Servers) getForwardConnection() (client net.Conn, error error) {
	for nodeName, node := range servers.config.Config.Nodes {
		if !node.IsController() {
			continue
		}

		apiServerAddress := fmt.Sprintf("%s:%d", node.IP, servers.config.Config.APIServerPort)

		client, error = net.Dial("tcp", apiServerAddress)
		if error == nil {
			return
		}

		log.WithFields(log.Fields{"name": nodeName, "address": apiServerAddress}).Error("node connection failed")
	}

	return
}

func (servers *Servers) forward(connection net.Conn) {
	client, error := servers.getForwardConnection()
	if error != nil {
		return
	}

	go func() {
		defer client.Close()
		defer connection.Close()
		io.Copy(client, connection)
	}()

	go func() {
		defer client.Close()
		defer connection.Close()
		io.Copy(connection, client)
	}()
}

func (servers *Servers) forwarder() error {
	listener, error := net.Listen("tcp", servers.config.GetForwarderAddress())
	if error != nil {
		return error
	}

	log.Info("started forwarder")

	go func() {
		for {
			connection, error := listener.Accept()
			if error != nil {
				log.WithFields(log.Fields{"error": error}).Error("forwarder accept failed")

				continue
			}

			go servers.forward(connection)
		}
	}()

	return nil
}

func (servers *Servers) Run() error {
	if error := servers.forwarder(); error != nil {
		return error
	}

	servers.config.Dump()

	for commandName, command := range servers.config.Config.Commands {
		if !config.CompareLabels(servers.config.Node.Labels, command.Labels) {
			continue
		}

		newCommand, error := servers.config.ApplyTemplate(commandName, command.Command)
		if error != nil {
			return error
		}

		go func() {
			for {
				// Keep trying until succeeding
				if _, error = utils.RunCommandWithOutput(newCommand); error == nil {
					log.WithFields(log.Fields{"command": newCommand}).Info("command executed")

					return
				}

				log.WithFields(log.Fields{"command": newCommand}).Error("command failed")

				time.Sleep(5 * time.Second)
			}
		}()
	}

	for name, serverConfig := range servers.config.Config.Servers {
		if !config.CompareLabels(servers.config.Node.Labels, serverConfig.Labels) {
			continue
		}

		server, error := NewServerWrapper(*servers.config, name, *serverConfig)

		if error != nil {
			return error
		}

		servers.add(server)
	}

	for _, server := range servers.servers {
		if error := server.Start(); error != nil {
			log.WithFields(log.Fields{"name": server.Name(), "error": error}).Error("server start failed")

			return error
		}

		log.WithFields(log.Fields{"name": server.Name()}).Info("started server")
	}

	defer func() {
		for _, server := range servers.servers {
			log.WithFields(log.Fields{"name": server.Name()}).Info("stopping server")

			server.Stop()
		}
	}()

	signals := make(chan os.Signal, 1)

	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	<-signals

	return nil
}
