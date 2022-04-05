package utils

import (
	"path"

	"github.com/darxkies/k8s-tew/data"
	log "github.com/sirupsen/logrus"
)

func GetTemplate(name string) string {
	content, error := data.Templates.ReadFile(path.Join("templates", name))
	if error != nil {
		log.WithFields(log.Fields{"name": name, "error": error}).Panic("Template failure")
	}

	return string(content)
}
