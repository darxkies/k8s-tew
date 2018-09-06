package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
)

const BINARY = "bin"
const LIBRARY = "lib"
const ALIASES = "aliases"
const LD_LINUX_PREFIX = "ld-linux-"

func help() {
	fmt.Println("Usage:")
	fmt.Printf("\t%s freeze <binary> <directory>\n", os.Args[0])
	fmt.Printf("\t%s deploy <source-directory> <destination-directory> <alias-directory>\n", os.Args[0])

	os.Exit(-1)
}

func fileExists(filename string) bool {
	_, error := os.Stat(filename)

	return !os.IsNotExist(error)
}

func copy(source, destination string) error {
	if !fileExists(source) {
		return nil
	}

	fmt.Printf("Copying %s to %s\n", source, destination)

	in, error := os.Open(source)
	if error != nil {
		return error
	}

	defer in.Close()

	out, error := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if error != nil {
		return error
	}

	defer out.Close()

	if _, error = io.Copy(out, in); error != nil {
		return error
	}

	return out.Sync()
}

func createDirectory(directoryName string) error {
	return os.MkdirAll(directoryName, os.ModePerm)
}

func copyBinary(binary, directory string) error {
	binaryPath := path.Join(directory, BINARY)

	if error := createDirectory(binaryPath); error != nil {
		return error
	}

	targetFilename := path.Join(binaryPath, path.Base(binary))

	return copy(binary, targetFilename)
}

func copyLibrary(library, directory string) error {
	libraryPath := path.Join(directory, LIBRARY)

	if error := createDirectory(libraryPath); error != nil {
		return error
	}

	targetFilename := path.Join(libraryPath, path.Base(library))

	return copy(library, targetFilename)
}

func freeze(binary, directory string) error {
	command := exec.Command("ldd", binary)

	output, error := command.CombinedOutput()
	if error != nil {
		return error
	}

	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		tokens := strings.Split(line, " ")

		if len(tokens) < 2 {
			continue
		}

		library := tokens[0]

		if len(tokens) == 4 {
			library = tokens[2]
		}

		library = strings.TrimSpace(library)

		copyLibrary(library, directory)
	}

	return copyBinary(binary, directory)
}

func copyDirectory(subdirectory, sourceDirectory, targetDirectory string) error {
	files, error := ioutil.ReadDir(path.Join(sourceDirectory, subdirectory))
	if error != nil {
		return error
	}

	sourceDirectory = path.Join(sourceDirectory, subdirectory)
	targetDirectory = path.Join(targetDirectory, subdirectory)

	if error := createDirectory(path.Join(targetDirectory)); error != nil {
		return error
	}

	for _, file := range files {
		sourceFilename := path.Join(sourceDirectory, file.Name())
		targetFilename := path.Join(targetDirectory, file.Name())

		if error := copy(sourceFilename, targetFilename); error != nil {
			return error
		}
	}

	return nil
}

func getLDLinuxLibrary(sourceDirectory string) (name string, error error) {
	files, error := ioutil.ReadDir(path.Join(sourceDirectory, LIBRARY))
	if error != nil {
		return
	}

	for _, file := range files {
		if strings.HasPrefix(file.Name(), LD_LINUX_PREFIX) {
			return file.Name(), nil
		}
	}

	return "", fmt.Errorf("ld-linux library not found")
}

func createAliases(sourceDirectory, targetDirectory, aliasDirectory string) error {
	files, error := ioutil.ReadDir(path.Join(sourceDirectory, BINARY))
	if error != nil {
		return error
	}

	ldLinuxLibrary, error := getLDLinuxLibrary(sourceDirectory)
	if error != nil {
		return error
	}

	ldLinuxLibrary = path.Join(aliasDirectory, LIBRARY, ldLinuxLibrary)

	fullAliasesFilename := path.Join(targetDirectory, ALIASES)

	result := ""

	for _, file := range files {
		result += fmt.Sprintf("alias %s=\"LD_LIBRARY_PATH=%s %s %s\"\n", file.Name(), path.Join(aliasDirectory, LIBRARY), ldLinuxLibrary, path.Join(aliasDirectory, BINARY, file.Name()))
	}

	fmt.Printf("Creating alias file %s\n", fullAliasesFilename)

	return ioutil.WriteFile(fullAliasesFilename, []byte(result), 0666)
}

func deploy(sourceDirectory, targetDirectory, aliasDirectory string) error {
	if error := copyDirectory(BINARY, sourceDirectory, targetDirectory); error != nil {
		return error
	}

	if error := copyDirectory(LIBRARY, sourceDirectory, targetDirectory); error != nil {
		return error
	}

	return createAliases(sourceDirectory, targetDirectory, aliasDirectory)
}

func main() {
	var error error

	if len(os.Args) == 4 && os.Args[1] == "freeze" {
		error = freeze(os.Args[2], os.Args[3])

	} else if len(os.Args) == 5 && os.Args[1] == "deploy" {
		error = deploy(os.Args[2], os.Args[3], os.Args[4])

	} else {
		help()
	}

	if error != nil {
		fmt.Println(error)

		os.Exit(-2)
	}
}
