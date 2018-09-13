package config

type Command struct {
	Name     string   `yaml:"name"`
	Command  string   `yaml:"command"`
	Labels   Labels   `yaml:"labels,omitempty"`
	Features Features `yaml:"features,omitempty"`
	OS       OS       `yaml:"os,omitempty"`
}

type Commands []*Command
type OS []string

func NewCommand(name string, labels Labels, features Features, os OS, command string) *Command {
	return &Command{Name: name, Labels: labels, Features: features, OS: os, Command: command}
}
