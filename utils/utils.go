package utils

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	oslib "github.com/redpois0n/goslib"
	log "github.com/sirupsen/logrus"
)

const COMMAND_TIMEOUT = 60 // In seconds

func WaitForSignal(signal <-chan struct{}, timeout uint) error {
	select {
	case <-signal:
		return nil

	case <-time.After(time.Duration(timeout) * time.Second):
		return errors.New("signal timeout")
	}
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
	_context, cancel := context.WithTimeout(context.Background(), COMMAND_TIMEOUT*time.Second)
	defer cancel()

	log.WithFields(log.Fields{"command": command}).Debug("Command started")

	cmd := exec.CommandContext(_context, "sh", "-c", command)

	output, error := cmd.CombinedOutput()
	if error != nil {
		log.WithFields(log.Fields{"command": command, "error": error}).Debug("Command failed")

		return "", fmt.Errorf("Command '%s' failed with error '%s' (Output: %s)", command, error, output)
	}

	log.WithFields(log.Fields{"command": command, "output": string(output)}).Debug("Command ended")

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

func ExtractImageName(value string) string {
	tokens := strings.Split(value, ":")

	if len(tokens) > 0 {
		return tokens[0]
	}

	return value
}

func ExtractImageTag(value string) string {
	tokens := strings.Split(value, ":")

	if len(tokens) > 1 {
		return tokens[1]
	}

	return value
}

func ApplyTemplate(label, content string, data interface{}, alternativeDelimiters bool) (string, error) {
	var result bytes.Buffer

	var functions = template.FuncMap{
		"unescape": func(value string) string {
			return value
		},
		"base64": func(value string) string {
			return base64.StdEncoding.EncodeToString([]byte(value))
		},
		"quoted_string_list": func(values []string) string {
			result := ""

			for i, value := range values {
				if i > 0 {
					result += ", "
				}

				result += "\"" + value + "\""
			}

			return result
		},
		"image_name": func(value string) string {
			return ExtractImageName(value)
		},
		"image_tag": func(value string) string {
			return ExtractImageTag(value)
		},
	}

	startDelimiter := "{{"
	endDelimiter := "}}"

	if alternativeDelimiters {
		startDelimiter = "[["
		endDelimiter = "]]"
	}

	argumentTemplate, error := template.New(label).Delims(startDelimiter, endDelimiter).Funcs(functions).Parse(content)
	if error != nil {
		return "", fmt.Errorf("Could not apply template '%s' (%s)", label, error.Error())
	}

	if error = argumentTemplate.Execute(&result, data); error != nil {
		return "", fmt.Errorf("Could not apply template '%s' (%s)", label, error.Error())
	}

	return result.String(), nil
}

func ApplyTemplateAndSave(label, templateName string, data interface{}, filename string, force bool, extendedDelimiters bool) error {
	content := GetTemplate(templateName)

	if FileExists(filename) && !force {
		LogFilename("Skipped", filename)

		return nil
	}

	content, error := ApplyTemplate(label, content, data, extendedDelimiters)
	if error != nil {
		return error
	}

	if error := ioutil.WriteFile(filename, []byte(content), 0644); error != nil {
		return fmt.Errorf("Could not write to '%s' (%s)", filename, error.Error())
	}

	LogFilename("Generated", filename)

	return nil
}

func GetBase64OfPEM(filename string) (string, error) {
	content, error := ioutil.ReadFile(filename)

	if error != nil {
		return "", fmt.Errorf("Could not read file '%s' (%s)", filename, error.Error())
	}

	return base64.StdEncoding.EncodeToString(content), nil
}

func GenerateCephKey() string {
	headerSize := 2 + 4 + 4 + 2
	keySize := 16
	buffer := make([]byte, headerSize+keySize)
	timestamp := time.Now().UnixNano()
	seconds := timestamp / 1000000000
	nanos := timestamp % 1000000000

	binary.LittleEndian.PutUint16(buffer[0:], 1)
	binary.LittleEndian.PutUint32(buffer[2:], uint32(seconds))
	binary.LittleEndian.PutUint32(buffer[6:], uint32(nanos))
	binary.LittleEndian.PutUint16(buffer[10:], uint16(keySize))

	rand.Read(buffer[headerSize:])

	return base64.StdEncoding.EncodeToString(buffer)
}

func GetOSName() string {
	return strings.ToLower(oslib.GetDist().Display)
}

func GetOSRelease() string {
	return oslib.GetDist().Release
}

func GetOSNameAndRelease() string {
	return fmt.Sprintf("%s/%s", GetOSName(), GetOSRelease())
}

func HasOS(os []string) bool {
	if len(os) == 0 {
		return true
	}

	for _, entry := range os {
		if entry == GetOSName() || entry == GetOSNameAndRelease() {
			return true
		}
	}

	return false
}
