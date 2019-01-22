package manifest

type HistoryEntryData struct {
	ID              string `json:"id"`
	Parent          string `json:"parent"`
	DockerVersion   string `json:"docker_version"`
	Architecture    string `json:"architecture"`
	OS              string `json:"os"`
	Container       string `json:"container"`
	Throwaway       bool   `json:"throwaway"`
	Config          Config `json:"config"`
	ContainerConfig Config `json:"container_config"`
	Created         string `json:"created"`
	Author          string `json:"author"`
}

type HistoryEntry struct {
	V1Compatibility string `json:"v1Compatibility"`
}

type History []*HistoryEntry
