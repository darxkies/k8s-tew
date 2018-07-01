package config

type AssetDirectory struct {
	Labels    Labels `yaml:"labels,omitempty"`
	Directory string `yaml:"directory"`
}

func NewAssetDirectory(labels []string, directory string) *AssetDirectory {
	return &AssetDirectory{Labels: labels, Directory: directory}
}
