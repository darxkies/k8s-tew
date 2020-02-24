package config

type Command struct {
	Name     string   `yaml:"name"`
	Command  string   `yaml:"command,omitempty"`
	Manifest string   `yaml:"command,omitempty"`
	Labels   Labels   `yaml:"labels,omitempty"`
	Features Features `yaml:"features,omitempty"`
	OS       OS       `yaml:"os,omitempty"`
}

type Commands []*Command
type OS []string

func NewCommand(name string, labels Labels, features Features, os OS, command string) *Command {
	return &Command{Name: name, Labels: labels, Features: features, OS: os, Command: command}
}

func NewManifest(name string, labels Labels, features Features, os OS, manifest string) *Command {
	return &Command{Name: name, Labels: labels, Features: features, OS: os, Manifest: manifest}
}
