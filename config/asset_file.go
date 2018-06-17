package config

type AssetFile struct {
	Labels    Labels `yaml:"labels,omitempty"`
	Directory string `yaml:"directory"`
}

func NewAssetFile(labels []string, directory string) *AssetFile {
	return &AssetFile{Labels: labels, Directory: directory}
}
