package servers

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/darxkies/k8s-tew/pkg/config"
	"github.com/darxkies/k8s-tew/pkg/utils"
)

type ServerWrapper struct {
	stop            bool
	name            string
	baseDirectory   string
	command         []string
	logger          config.LoggerConfig
	pathEnvironment string
	started         bool
	context         context.Context
	cancel          context.CancelFunc
	done            chan bool
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
	if server.started {
		return fmt.Errorf("%s already started", server.name)
	}

	server.stop = false

	if server.logger.Enabled {
		logsDirectory := filepath.Dir(server.logger.Filename)

		if error := utils.CreateDirectoryIfMissing(logsDirectory); error != nil {
			return error
		}
	}

	log.WithFields(log.Fields{"name": server.Name(), "_command": strings.Join(server.command, " ")}).Info("Starting server")

	server.context, server.cancel = context.WithCancel(context.Background())
	server.done = make(chan bool, 1)

	server.started = true

	go func() {
		for !server.stop {
			command := exec.CommandContext(server.context, server.command[0], server.command[1:]...)
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

			error = command.Run()

			if !server.stop {
				time.Sleep(time.Second)

				log.WithFields(log.Fields{"name": server.name, "error": error, "_command": strings.Join(server.command, " ")}).Error("Restarting server")
			}
		}

		close(server.done)
	}()

	return nil
}

func (server *ServerWrapper) Stop() {
	if !server.started {
		return
	}

	server.stop = true

	server.cancel()

	<-server.done

	log.WithFields(log.Fields{"name": server.name, "_command": strings.Join(server.command, " ")}).Info("Stopped server")
}

func (server *ServerWrapper) Name() string {
	return server.name
}
