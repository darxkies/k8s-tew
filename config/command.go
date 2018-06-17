package config

type Command struct {
	Name    string
	Command string
	Labels  Labels
}

type Commands []*Command

func NewCommand(name string, labels Labels, command string) *Command {
	return &Command{Name: name, Labels: labels, Command: command}
}
