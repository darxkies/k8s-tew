package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"syscall"
)

const _freezer = "freezer"
const _binary = "bin"
const _library = "lib"
const _ldLinuxPrefix = "ld-linux-"

// Print usage help and exit
func help() {
	log.Println("Usage:")
	log.Printf("\t%s freeze-binary <binary> <directory>\n", os.Args[0])
	log.Printf("\t%s freeze-library <library> <directory>\n", os.Args[0])
	log.Printf("\t%s deploy <source-directory> <destination-directory>\n", os.Args[0])

	os.Exit(-1)
}

// Check if file exists
func fileExists(filename string) bool {
	_, error := os.Stat(filename)

	return !os.IsNotExist(error)
}

// Copy the content of one file to another
func copy(source, destination string) error {
	// Skip if the destination file exists
	if fileExists(destination) {
		return nil
	}

	// Exit if the file does not exist
	if !fileExists(source) {
		return nil
	}

	log.Printf("Copying %s to %s\n", source, destination)

	// Open source file
	in, error := os.Open(source)
	if error != nil {
		return error
	}

	// Defer source file closing
	defer in.Close()

	// Open target file
	out, error := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if error != nil {
		return error
	}

	// Defer target file closing
	defer out.Close()

	// Copy file content
	if _, error = io.Copy(out, in); error != nil {
		return error
	}

	// Sync content to storage
	return out.Sync()
}

// Crate a directory
func createDirectory(directoryName string) error {
	return os.MkdirAll(directoryName, os.ModePerm)
}

// Copy an executable to the binary folder
func copyBinary(binary, directory string) error {
	binaryPath := path.Join(directory, _binary)

	// Create binary folder if it does not exist
	if error := createDirectory(binaryPath); error != nil {
		return error
	}

	// Generate full filename
	targetFilename := path.Join(binaryPath, path.Base(binary))

	// Copy file content
	return copy(binary, targetFilename)
}

// Copy a library to the library directory
func copyLibrary(library, directory string) error {
	libraryPath := path.Join(directory, _library)

	// Create library folder if missing
	if error := createDirectory(libraryPath); error != nil {
		return error
	}

	// Create full library filename
	targetFilename := path.Join(libraryPath, path.Base(library))

	// Copy library content
	return copy(library, targetFilename)
}

func freezeDependencies(filename, directory string) error {
	// Generate ldd command to extract library information related to the file
	command := exec.Command("ldd", filename)

	// Get ldd output
	output, error := command.CombinedOutput()
	if error != nil {
		return error
	}

	// Split output to lines
	lines := strings.Split(string(output), "\n")

	// Iterate over all lines
	for _, line := range lines {
		// Split line into tokens
		tokens := strings.Split(line, " ")

		// Skip line if too short
		if len(tokens) < 2 {
			continue
		}

		// Get library name
		library := tokens[0]

		// If the line is long get the right library name
		if len(tokens) == 4 {
			library = tokens[2]
		}

		// Remove leading and trailing spaces
		library = strings.TrimSpace(library)

		// Copy library
		if error := copyLibrary(library, directory); error != nil {
			return error
		}
	}

	return nil
}

// Copy library and its dependencies
func freezeLibrary(library, directory string) error {
	// Extract dependencies
	if error := freezeDependencies(library, directory); error != nil {
		return error
	}

	// Copy library
	return copyLibrary(library, directory)
}

// Copy binary, its libraries and generate wrapper for the binary
func freezeBinary(freezerExecutable, binary, directory string) error {
	// Generate wrapper binary filename
	wrapperBinary := path.Join(directory, path.Base(binary))

	// Bail out if it already exists
	if fileExists(wrapperBinary) {
		return nil
	}

	// Extract dependencies
	if error := freezeDependencies(binary, directory); error != nil {
		return error
	}

	// Copy binary
	if error := copyBinary(binary, directory); error != nil {
		return error
	}

	// Generate wrapper
	return copy(freezerExecutable, wrapperBinary)
}

// Copy the content of a directory to a new directory
func copyDirectory(subdirectory, sourceDirectory, targetDirectory string) error {
	// Generate full source and target directory names
	sourceDirectory = path.Join(sourceDirectory, subdirectory)
	targetDirectory = path.Join(targetDirectory, subdirectory)

	// Get content of source directory
	files, error := ioutil.ReadDir(sourceDirectory)
	if error != nil {
		return error
	}

	// Crate target directory
	if error := createDirectory(targetDirectory); error != nil {
		return error
	}

	// Iterate over all files in source directory
	for _, file := range files {
		// Generate full source and target filename
		sourceFilename := path.Join(sourceDirectory, file.Name())
		targetFilename := path.Join(targetDirectory, file.Name())

		// Copy file to new directry
		if error := copy(sourceFilename, targetFilename); error != nil {
			return error
		}
	}

	return nil
}

// Retrieve ld-linux-....so filename from the library directory
func getLDLinuxLibrary(libraryDirectory string) (name string, error error) {
	// Get the names of all files in the library directory
	files, error := ioutil.ReadDir(libraryDirectory)
	if error != nil {
		return
	}

	// Iterate over all files in the library directory
	for _, file := range files {
		// Check if it is the wanted file and exit if true
		if strings.HasPrefix(file.Name(), _ldLinuxPrefix) {
			return file.Name(), nil
		}
	}

	// Return an error
	return "", fmt.Errorf("ld-linux library not found")
}

// Copy binary and library files to a new directory
func deploy(sourceDirectory, targetDirectory string) error {
	// Copy binary files
	if error := copyDirectory(_binary, sourceDirectory, targetDirectory); error != nil {
		return error
	}

	// Copy library files
	if error := copyDirectory(_library, sourceDirectory, targetDirectory); error != nil {
		return error
	}

	// Copy wrapper files
	return copyDirectory("", sourceDirectory, targetDirectory)
}

// Run a frozen binary by setting environment variables and fixing the arguments
func wrapper(freezerExecutable string, arguments []string) error {
	// At least the name of the executable has to be provided
	if len(arguments) < 1 {
		return fmt.Errorf("Not enough arguments")
	}

	// Get Executable Path
	executablePath, error := filepath.Abs(filepath.Dir(freezerExecutable))
	if error != nil {
		return error
	}

	// Get Executable Name
	executableName := path.Base(arguments[0])

	// Patch full executable name to point to the real binary
	arguments[0] = path.Join(executablePath, _binary, executableName)

	// Get library directory
	ldLibraryPath := path.Join(executablePath, _library)

	// If there is a ld-linux...so find it and prepend it to the list of arguments in front of the executable itself
	ldLinuxName, error := getLDLinuxLibrary(ldLibraryPath)
	if error == nil {
		ldLinuxName = path.Join(ldLibraryPath, ldLinuxName)

		arguments = append([]string{ldLinuxName}, arguments...)
	}

	// Create environment to execute the command
	command := exec.Command(arguments[0], arguments[1:]...)

	// Create environment and point to the library directory
	command.Env = os.Environ()
	command.Env = append(command.Env, fmt.Sprintf("LD_LIBRARY_PATH=%s", ldLibraryPath))

	// Redirect every stdout, stderr and stdin to the caller
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin

	// Start command
	if error := command.Start(); error != nil {
		return error
	}

	// Check the exit code of the executable and if not 0 exit our program with the same exit code
	if error := command.Wait(); error != nil {
		if exitError, ok := error.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				os.Exit(status.ExitStatus())
			}
		} else {
			return error
		}
	}

	return nil
}

func main() {
	var error error

	// Get full executable filename
	freezerExecutable, error := os.Executable()
	if error != nil {
		log.Println(error)

		os.Exit(-3)
	}

	// Copy binary and its libraries to the specified directory
	if len(os.Args) == 4 && os.Args[1] == "freeze-binary" {
		error = freezeBinary(freezerExecutable, os.Args[2], os.Args[3])

		// Copy library and its dependencies to the specified directory
	} else if len(os.Args) == 4 && os.Args[1] == "freeze-library" {
		error = freezeLibrary(os.Args[2], os.Args[3])

		// Copy bin and lib foders to the target directory
	} else if len(os.Args) == 5 && os.Args[1] == "deploy" {
		error = deploy(os.Args[2], os.Args[3])

		// Execute the binary from the bin folder that has the same name as this executable
	} else if path.Base(os.Args[0]) != _freezer {
		error = wrapper(freezerExecutable, os.Args)

		// Print usage information
	} else {
		help()
	}

	// Print error message if any
	if error != nil {
		log.Println(error)

		os.Exit(-2)
	}
}
