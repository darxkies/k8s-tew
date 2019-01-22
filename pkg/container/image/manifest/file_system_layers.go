package manifest

const EmptyLayer = "sha256:a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4"

type FileSystemLayer struct {
	BlobSum string `json:"blobSum"`
}

type FileSystemLayers []FileSystemLayer
