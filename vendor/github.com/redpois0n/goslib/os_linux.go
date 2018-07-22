package oslib

const (
	Name = "Linux"
)

func GetDisplay() string {
	distro := GetDist()

	if distro.Display != "" {
		return distro.Display + " " + distro.Release
	}

	return Name
}

func GetVersion() string {
	return ""
}
