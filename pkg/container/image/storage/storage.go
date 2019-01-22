package storage

import (
	"archive/tar"
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"
)

type Storage interface {
	WriteFile(filename string, data []byte) error
	Close() error
	Remove() error
}

type TarStorage struct {
	filename  string
	tarFile   *os.File
	tarBall   *tar.Writer
	timestamp time.Time
}

func NewTarStorage(filename string) (*TarStorage, error) {
	result := &TarStorage{filename: filename, timestamp: time.Now()}

	var error error

	result.tarFile, error = os.Create(filename)
	if error != nil {
		return nil, errors.Wrapf(error, "could not open %s", filename)
	}

	result.tarBall = tar.NewWriter(result.tarFile)

	return result, nil
}

func (storage *TarStorage) WriteFile(filename string, data []byte) error {
	tarHeader := new(tar.Header)
	tarHeader.Name = filename
	tarHeader.Size = int64(len(data))
	tarHeader.Mode = 0660
	tarHeader.ModTime = storage.timestamp

	if error := storage.tarBall.WriteHeader(tarHeader); error != nil {
		return errors.Wrapf(error, "could not write header for %s to %s", filename, storage.filename)
	}

	written, error := storage.tarBall.Write(data)

	if written != len(data) {
		return fmt.Errorf("could not write all data for %s to %s", filename, storage.filename)
	}

	if error != nil {
		return errors.Wrapf(error, "could not write data for %s to %s", filename, storage.filename)
	}

	return nil
}

func (storage *TarStorage) Close() error {
	if error := storage.tarBall.Close(); error != nil {
		return errors.Wrapf(error, "could not close tarball for %s", storage.filename)
	}

	if error := storage.tarFile.Close(); error != nil {
		return errors.Wrapf(error, "could not close tarfile for %s", storage.filename)
	}

	return nil
}

func (storage *TarStorage) Remove() error {
	return os.Remove(storage.filename)
}
