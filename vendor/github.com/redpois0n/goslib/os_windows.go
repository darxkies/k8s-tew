package oslib

import (
	"bytes"
	"os/exec"
	"strings"
)

const (
	Name = "Windows"
)

func GetVersion() string {
	cmd := exec.Command("cmd")

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		panic(err)
	}

	raw := out.String()
	i1 := strings.Index(raw, "[Version")
	i2 := strings.Index(raw, "]")
	var ver string

	if i1 == -1 || i2 == -1 {
		ver = ""
	} else {
		ver = raw[i1+len("[Version") : i2]
	}

	return strings.Trim(ver, " ")
}

func getEdition() string {
	version := GetVersion()
	parts := strings.Split(version, ".")
	majormin := parts[0] + "." + parts[1]

	var edition string

	switch majormin {
	case "10.0": // 10 Server
		edition = "10"
	case "6.3": // Server 2012 R2
		edition = "8.1"
	case "6.2": // Server 2012
		edition = "8"
	case "6.1":
		edition = "7"
	case "6.0":
		edition = "Vista"
	case "5.2":
		edition = "Server 2003"
	case "5.1":
		edition = "XP"
	case "5.0":
		edition = "2000"
	}

	return edition
}

func GetDisplay() string {
	display := Name
	version := getEdition()

	if version != "" {
		display += " " + version
	}

	return display
}
