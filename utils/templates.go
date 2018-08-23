package utils

import (
	"github.com/gobuffalo/packr"
	log "github.com/sirupsen/logrus"
)

var box packr.Box

func init() {
	box = packr.NewBox("../templates")
}

func GetTemplate(name string) string {
	content, error := box.MustString(name)

	if error != nil {
		log.WithFields(log.Fields{"name": name, "error": error}).Panic("Template failure")
	}

	return content
}
