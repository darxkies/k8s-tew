package config

type LoggerConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Filename string `yaml:"filename"`
}
