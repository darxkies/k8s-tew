package config

type AssetFile struct {
	Labels    Labels `yaml:"labels,omitempty"`
	Filename  string `yaml:"filename,omitempty"`
	Directory string `yaml:"directory"`
}

func NewAssetFile(labels []string, filename, directory string) *AssetFile {
	return &AssetFile{Labels: labels, Filename: filename, Directory: directory}
}
