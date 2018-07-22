//+build freebsd,openbsd,dragonfly,netbsd

package oslib

func GetVersion() string {
	return ""
}

func GetDisplay() string {
	return Name
}
