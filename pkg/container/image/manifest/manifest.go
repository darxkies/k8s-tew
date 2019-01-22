package manifest

type Manifest struct {
	SchemaVersion    int              `json:"schemaVersion"`
	Name             string           `json:"name"`
	Tag              string           `json:"tag"`
	Architecture     string           `json:"architecture"`
	FileSystemLayers FileSystemLayers `json:"fsLayers"`
	History          History          `json:"history"`
}
