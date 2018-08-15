package utils

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	oslib "github.com/redpois0n/goslib"
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
		"base64": func(value string) template.HTML {
			return template.HTML(base64.StdEncoding.EncodeToString([]byte(value)))
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

var _spinner *spinner.Spinner
var _progressSteps int
var _progressStep int
var _progressShow bool

func init() {
	_spinner = spinner.New(spinner.CharSets[9], 100*time.Millisecond)
}

func ShowProgress() {
	_spinner.Prefix = "["
	_spinner.Suffix = fmt.Sprintf("] Progress: %d/%d", _progressStep+1, _progressSteps)

	_spinner.Start()

	_progressShow = true
}

func HideProgress() {
	_spinner.Stop()

	_progressShow = false
}

func IncreaseProgressStep() {
	_progressStep += 1
}

func SetProgressSteps(steps int) {
	_progressSteps = steps
}

type logrusHook struct{}

func (hook logrusHook) Fire(entry *log.Entry) error {
	show := _progressShow

	HideProgress()

	entry.Message = fmt.Sprintf("[%d/%d]", _progressStep+1, _progressSteps) + " " + entry.Message

	if show {
		ShowProgress()
	}

	return nil
}

func (hook logrusHook) Levels() []log.Level {
	return log.AllLevels
}

func SetupLogger() {
	log.AddHook(logrusHook{})
}
