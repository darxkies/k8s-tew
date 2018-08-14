package config

type AssetDirectory struct {
	Labels    Labels `yaml:"labels,omitempty"`
	Directory string `yaml:"directory"`
	Absolute  bool   `yaml:"absolute"`
}

func NewAssetDirectory(labels []string, directory string, absolute bool) *AssetDirectory {
	return &AssetDirectory{Labels: labels, Directory: directory, Absolute: absolute}
}
