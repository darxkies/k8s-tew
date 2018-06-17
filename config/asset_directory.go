package config

type AssetDirectory struct {
	Directory string `yaml:"directory"`
}

func NewAssetDirectory(directory string) *AssetDirectory {
	return &AssetDirectory{Directory: directory}
}
