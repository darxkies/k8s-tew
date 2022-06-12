package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Checksum struct {
	value   string
	updated bool
}
type Checksums struct {
	filename      string
	baseDirectory string
	checksums     map[string]Checksum
	loaded        bool
}

func NewChecksums(filename, baseDirectory string) *Checksums {
	return &Checksums{filename: filename, baseDirectory: baseDirectory, checksums: map[string]Checksum{}}
}

func (checksums *Checksums) md5(targetFilename string) (result string, error error) {
	file, error := os.Open(targetFilename)
	if error != nil {
		return
	}

	defer file.Close()

	hash := md5.New()

	if _, error = io.Copy(hash, file); error != nil {
		return
	}

	result = hex.EncodeToString(hash.Sum(nil)[:16])

	return
}

func (checksums *Checksums) GetChecksum(targetFilename string) (result string, error error) {
	relativeFilename := targetFilename[len(checksums.baseDirectory):]

	checksumsCache, _errorCache := os.Stat(checksums.filename)
	checksumsTarget, _errorTarget := os.Stat(targetFilename)

	if _errorCache == nil && _errorTarget == nil && checksumsTarget.ModTime().Before(checksumsCache.ModTime()) {
		if checksum, ok := checksums.checksums[relativeFilename]; ok {
			return checksum.value, nil
		}
	}

	if checksum, ok := checksums.checksums[relativeFilename]; ok && checksum.updated {
		return checksum.value, nil
	}

	result, _error := checksums.md5(targetFilename)
	if _error != nil {
		return "", _error
	}

	log.WithFields(log.Fields{"file": targetFilename, "checksum": result}).Debug("Updated checksum")

	checksums.checksums[relativeFilename] = Checksum{value: result, updated: true}

	return
}

func (checksums *Checksums) Save() error {
	buffer := ""

	for filename, value := range checksums.checksums {
		buffer += fmt.Sprintf("%s %s\n", value.value, filename)
	}

	if _error := ioutil.WriteFile(checksums.filename, []byte(buffer), 0644); _error != nil {
		return errors.Wrapf(_error, "Could not write to %s", checksums.filename)
	}

	return nil
}

func (checksums *Checksums) Load() error {
	if checksums.loaded {
		return nil
	}

	checksums.loaded = true

	content, _error := ioutil.ReadFile(checksums.filename)
	if _error != nil {
		return _error
	}

	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		if len(line) < 35 {
			continue
		}

		filename := line[33:]
		checksum := line[:32]

		checksums.checksums[filename] = Checksum{value: checksum, updated: false}
	}

	return nil
}
