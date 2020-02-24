package utils

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	oslib "github.com/wille/osutil"
)

const commandTimeout = 60 // In seconds

// WaitForSignal exists when a signal was fired or a timeout occurred
func WaitForSignal(signal <-chan struct{}, timeout uint) error {
	select {
	case <-signal:
		return nil

	case <-time.After(time.Duration(timeout) * time.Second):
		return errors.New("signal timeout")
	}
}

// GetWorkingDirectory returns the working directory of the executable
func GetWorkingDirectory() (string, error) {
	return os.Getwd()
}

// CreateDirectoryIfMissing creates a directory if it does not exist
func CreateDirectoryIfMissing(directoryName string) error {
	if stat, error := os.Stat(directoryName); error == nil && !stat.IsDir() {
		return fmt.Errorf("'%s' already exists but it is not a directory", directoryName)
	}

	return os.MkdirAll(directoryName, os.ModePerm)
}

// CreateFileIfMissing writes a string to a file
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

// FileExists returns true if a file exists
func FileExists(filename string) bool {
	_, error := os.Stat(filename)

	return !os.IsNotExist(error)
}

// RunCommandWithOutput execute a shell command and return its output
func RunCommandWithOutput(command string) (string, error) {
	_context, cancel := context.WithTimeout(context.Background(), commandTimeout*time.Second)
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

// RunCommand executes a shell command
func RunCommand(command string) error {
	_, error := RunCommandWithOutput(command)

	return error
}

// GetURL assembles a URL
func GetURL(protocol, ip string, port uint16) string {
	return fmt.Sprintf("%s://%s:%d", protocol, ip, port)
}

// OpenWebBrowser starts a web browser
func OpenWebBrowser(name, url string) error {
	if _, error := RunCommandWithOutput(fmt.Sprintf("xdg-open %s", url)); error != nil {
		return fmt.Errorf("Could not open %s at %s (%s)", name, url, error.Error())
	}

	return nil
}

// IsRoot returns true if the program is executed with root rights
func IsRoot() bool {
	return os.Geteuid() == 0
}

// ExtractImageName returns the name of a Docker image
func ExtractImageName(value string) string {
	tokens := strings.Split(value, ":")

	if len(tokens) > 0 {
		return tokens[0]
	}

	return value
}

// ExtractImageTag returns the tag from a Docker image
func ExtractImageTag(value string) string {
	tokens := strings.Split(value, ":")

	if len(tokens) > 1 {
		return tokens[1]
	}

	return value
}

// ApplyTemplate generates a string using a template
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

// ApplyTemplateAndSave generates the content of a file based on a template
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

// GetBase64OfPEM reads the content of a PEM file and converts it to Base64
func GetBase64OfPEM(filename string) (string, error) {
	content, error := ioutil.ReadFile(filename)

	if error != nil {
		return "", fmt.Errorf("Could not read file '%s' (%s)", filename, error.Error())
	}

	return base64.StdEncoding.EncodeToString(content), nil
}

// GenerateCephKey returns a valid ceph key
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

	_, _ = rand.Read(buffer[headerSize:])

	return base64.StdEncoding.EncodeToString(buffer)
}

// GetOSName returns the name of the operating system
func GetOSName() string {
	return strings.ToLower(oslib.GetDist().Display)
}

// GetOSRelease returns the version of the operating system
func GetOSRelease() string {
	return oslib.GetDist().Release
}

// GetOSNameAndRelease returns the name of the operating system and the operating system version
func GetOSNameAndRelease() string {
	return fmt.Sprintf("%s/%s", GetOSName(), GetOSRelease())
}

// HasOS checks if parameter os contains the name of the current operating system
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

// MoveFile copies a files and then removes the original
func MoveFile(sourceFilename, targetFilename string) error {
	{
		sourceHandle, error := os.Open(sourceFilename)
		if error != nil {
			return errors.Wrapf(error, "Could not open source source file %s", sourceFilename)
		}

		defer sourceHandle.Close()

		targetHandle, error := os.OpenFile(targetFilename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0555)
		if error != nil {
			return errors.Wrapf(error, "Could not open target file %s", targetFilename)
		}

		defer targetHandle.Close()

		_, error = io.Copy(targetHandle, sourceHandle)
		if error != nil {
			return errors.Wrapf(error, "Could not write to target file %s", targetFilename)
		}
	}

	if error := os.Remove(sourceFilename); error != nil {
		return errors.Wrapf(error, "Could not remove source file %s", sourceFilename)
	}

	return nil
}
