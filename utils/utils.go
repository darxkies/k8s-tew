package utils

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
)

func WaitForSignal(signal <-chan struct{}, timeout uint) error {
	select {
	case <-signal:
		return nil

	case <-time.After(time.Duration(timeout) * time.Second):
		return errors.New("signal timeout")
	}

	return nil
}

func GetWorkingDirectory() (string, error) {
	return os.Getwd()
}

func CreateDirectoryIfMissing(directoryName string) error {
	if stat, error := os.Stat(directoryName); error == nil && !stat.IsDir() {
		return fmt.Errorf("'%s' already exists but it is not a directory.", directoryName)
	}

	return os.MkdirAll(directoryName, os.ModePerm)
}

func CreateFileIfMissing(filename, content string) error {
	if _, error := os.Stat(filename); !os.IsNotExist(error) {
		return nil
	}

	directoryName := filepath.Dir(filename)

	if error := CreateDirectoryIfMissing(directoryName); error != nil {
		return error
	}

	return ioutil.WriteFile(filename, []byte(content), 0644)
}

func FileExists(filename string) bool {
	_, error := os.Stat(filename)

	return !os.IsNotExist(error)
}

func RunCommandWithOutput(command string) (string, error) {
	cmd := exec.Command("sh", "-c", command)

	output, error := cmd.CombinedOutput()

	if error != nil {
		return "", errors.New(fmt.Sprintf("Command '%s' failed with error '%s' (Output: %s)", command, error, output))
	}

	return string(output), nil
}

func RunCommand(command string) error {
	_, error := RunCommandWithOutput(command)

	return error
}

func RunSSHClient(ip string) {
	command := fmt.Sprintf("ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -t ubuntu@%s \"sudo su -\"", ip)

	cmd := exec.Command("sh", "-c", command)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	_ = cmd.Run()
}

func IsRoot() bool {
	return os.Geteuid() == 0
}

func ApplyTemplate(content string, data interface{}) (string, error) {
	var result bytes.Buffer

	var functions = template.FuncMap{
		"unescape": func(value string) template.HTML {
			return template.HTML(value)
		},
		"quoted_string_list": func(values []string) template.HTML {
			result := ""

			for i, value := range values {
				if i > 0 {
					result += ", "
				}

				result += "\"" + value + "\""
			}

			return template.HTML(result)
		},
	}

	argumentTemplate, error := template.New("ApplyTemplate").Funcs(functions).Parse(content)
	if error != nil {
		return "", error
	}

	if error = argumentTemplate.Execute(&result, data); error != nil {
		return "", error
	}

	return result.String(), nil
}

func ApplyTemplateAndSave(content string, data interface{}, filename string, force bool) error {
	if FileExists(filename) && !force {
		log.WithFields(log.Fields{"filename": filename}).Info("skipped")

		return nil
	}

	content, error := ApplyTemplate(content, data)
	if error != nil {
		return error
	}

	if error := ioutil.WriteFile(filename, []byte(content), 0644); error != nil {
		return error
	}

	log.WithFields(log.Fields{"filename": filename}).Info("generated")

	return nil
}

func GetBase64OfPEM(filename string) (string, error) {
	content, error := ioutil.ReadFile(filename)

	if error != nil {
		return "", error
	}

	return base64.StdEncoding.EncodeToString(content), nil
}
