package servers

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/darxkies/k8s-tew/config"
	"github.com/darxkies/k8s-tew/utils"

	log "github.com/sirupsen/logrus"
)

type ServerWrapper struct {
	stop          bool
	name          string
	baseDirectory string
	command       []string
}

func NewServerWrapper(_config config.InternalConfig, name string, serverConfig config.ServerConfig) (Server, error) {
	var error error

	serverConfig.Command, error = _config.ApplyTemplate("command", serverConfig.Command)

	if error != nil {
		return nil, error
	}

	server := &ServerWrapper{name: name, baseDirectory: _config.BaseDirectory, command: []string{serverConfig.Command}}

	for key, value := range serverConfig.Arguments {
		if len(value) == 0 {
			server.command = append(server.command, fmt.Sprintf("--%s", key))

		} else {
			newValue, error := _config.ApplyTemplate(fmt.Sprintf("%s.%s", server.Name, key), value)

			if error != nil {
				return nil, error
			}

			server.command = append(server.command, fmt.Sprintf("--%s=%s", key, newValue))
		}
	}

	return server, nil
}

func (server *ServerWrapper) Start() error {
	server.stop = false

	criDirectory := path.Join(server.baseDirectory, utils.GetFullCRIBinariesDirectory())
	logsDirectory := path.Join(server.baseDirectory, utils.GetFullLoggingDirectory())

	if error := utils.CreateDirectoryIfMissing(logsDirectory); error != nil {
		return error
	}

	logFilename := path.Join(logsDirectory, server.name+".log")

	go func() {
		for !server.stop {
			logFile, error := os.OpenFile(logFilename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

			if error != nil {
				log.WithFields(log.Fields{"filename": logFile, "error": error}).Error("could no open file")

				continue
			}

			defer logFile.Close()

			command := exec.Command(server.command[0], server.command[1:]...)

			os.Setenv("PATH", fmt.Sprintf("%s:%s", criDirectory, os.Getenv("PATH")))

			command.Stdout = logFile
			command.Stderr = logFile
			command.Run()

			time.Sleep(time.Second)

			log.WithFields(log.Fields{"name": server.name}).Error("server terminated")
		}
	}()

	return nil
}

func (server *ServerWrapper) Stop() {
	server.stop = true
}

func (server *ServerWrapper) Name() string {
	return server.name
}
