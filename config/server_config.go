package config

import (
	log "github.com/sirupsen/logrus"
)

type ServerConfig struct {
	Name      string            `yaml:"name"`
	Labels    Labels            `yaml:"labels"`
	Logger    LoggerConfig      `yaml:"logger"`
	Command   string            `yaml:"command"`
	Arguments map[string]string `yaml:"arguments"`
}

type Servers []ServerConfig

func (config ServerConfig) Dump() {
	log.WithFields(log.Fields{"name": config.Name, "labels": config.Labels, "command": config.Command}).Info("config server")

	for key, value := range config.Arguments {
		log.WithFields(log.Fields{"name": config.Name, "argument": key, "value": value}).Info("config server argument")
	}
}
