package config

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"

	"github.com/darxkies/k8s-tew/utils"
)

func GetBase64OfPEM(filename string) (string, error) {
	content, error := ioutil.ReadFile(filename)

	if error != nil {
		return "", error
	}

	return base64.StdEncoding.EncodeToString(content), nil
}

func GenerateConfigKubeConfig(kubeConfigFilename, caFilename, user, apiServers, certificateFilename, keyFilename string) error {
	base64CA, error := GetBase64OfPEM(caFilename)

	if error != nil {
		return error
	}

	base64Certificate, error := GetBase64OfPEM(certificateFilename)

	if error != nil {
		return error
	}

	base64Key, error := GetBase64OfPEM(keyFilename)

	if error != nil {
		return error
	}

	result := fmt.Sprintf(utils.KUBE_CONFIG_TEMPLATE, base64CA, apiServers, user, base64Certificate, base64Key, user)

	return ioutil.WriteFile(kubeConfigFilename, []byte(result), 0644)
}
