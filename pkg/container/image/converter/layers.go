package converter

type layer struct {
	MediaType  string
	Size       int64
	BlobDigest string
	TarDigest  string
	Filename   string
}

type layers []*layer
