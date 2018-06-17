package config

type AssetConfig struct {
	Directories map[string]*AssetDirectory
	Files       map[string]*AssetFile
}
