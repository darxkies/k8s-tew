package utils

import (
	"io"

	"github.com/gobuffalo/packr"
	log "github.com/sirupsen/logrus"
)

var templatesBox packr.Box
var embeddedBox packr.Box

func init() {
	templatesBox = packr.NewBox("../templates")
	embeddedBox = packr.NewBox("../embedded")
}

func GetTemplate(name string) string {
	content, error := templatesBox.MustString(name)

	if error != nil {
		log.WithFields(log.Fields{"name": name, "error": error}).Panic("Template failure")
	}

	return content
}

func GetEmbeddedFiles(callback func(path string, readCloser io.ReadCloser) error) error {
	return embeddedBox.Walk(func(path string, file packr.File) error {
		return callback(path, file)
	})
}
