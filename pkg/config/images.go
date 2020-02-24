package config

import (
	"crypto/sha256"
	"fmt"
)

type Image struct {
	Name     string
	Features Features
}

type Images []Image

func (image Image) GetImageFilename() (result string) {

	for _, char := range image.Name {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') {
			result += string(char)

			continue
		}

		result += "_"
	}

	result = fmt.Sprintf("%s.%X.tar", result, sha256.Sum256([]byte(image.Name)))

	return
}
