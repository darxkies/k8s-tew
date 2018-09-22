package servers

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/darxkies/k8s-tew/config"
	"github.com/darxkies/k8s-tew/utils"

	log "github.com/sirupsen/logrus"
)

type ServerWrapper struct {
	stop            bool
	name            string
	baseDirectory   string
	command         []string
	logger          config.LoggerConfig
	pathEnvironment string
}

func NewServerWrapper(_config config.InternalConfig, name string, serverConfig config.ServerConfig, pathEnvironment string) (Server, error) {
	var error error

	serverConfig.Command, error = _config.ApplyTemplate("command", serverConfig.Command)

	if error != nil {
		return nil, error
	}

	server := &ServerWrapper{name: name, baseDirectory: _config.BaseDirectory, command: []string{serverConfig.Command}, logger: serverConfig.Logger, pathEnvironment: pathEnvironment}

	server.logger.Filename, error = _config.ApplyTemplate("LoggingDirectory", server.logger.Filename)
	if error != nil {
		return nil, error
	}

	for key, value := range serverConfig.Arguments {
		if len(value) == 0 {
			server.command = append(server.command, fmt.Sprintf("--%s", key))

		} else {
			newValue, error := _config.ApplyTemplate(fmt.Sprintf("%s.%s", server.Name(), key), value)
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

	if server.logger.Enabled {
		logsDirectory := filepath.Dir(server.logger.Filename)

		if error := utils.CreateDirectoryIfMissing(logsDirectory); error != nil {
			return error
		}
	}

	log.WithFields(log.Fields{"name": server.Name(), "_command": strings.Join(server.command, " ")}).Info("Starting server")

	go func() {
		for !server.stop {
			command := exec.Command(server.command[0], server.command[1:]...)
			command.SysProcAttr = &syscall.SysProcAttr{
				Setpgid: true,
				Pgid:    0,
			}

			command.Env = os.Environ()
			command.Env = append(command.Env, server.pathEnvironment)

			var logFile *os.File
			var error error

			if server.logger.Enabled {
				logFile, error = os.OpenFile(server.logger.Filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

				if error != nil {
					log.WithFields(log.Fields{"filename": logFile, "error": error}).Error("Could not open file")

					continue
				}

				command.Stdout = logFile
				command.Stderr = logFile
			}

			defer func() {
				if logFile != nil {
					logFile.Close()
				}
			}()

			command.Run()

			time.Sleep(time.Second)

			if !server.stop {
				log.WithFields(log.Fields{"name": server.name, "_command": strings.Join(server.command, " ")}).Error("Restarting server")
			}
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
