package converter

import "github.com/darxkies/k8s-tew/pkg/container/image/manifest"

type layer struct {
	MediaType  string
	Size       int64
	BlobDigest string
	TarDigest  string
	Filename   string
	EmptyLayer bool
	History    *manifest.HistoryEntryData
}

type layers []*layer
