package oslib

/*
#include <sys/sysctl.h>
*/
import "C"

import (
	"strconv"
	"strings"
	"unsafe"
)

const (
	Name = "macOS"
)

type version struct {
	release, name, releaseName string
}

var mapping map[int]version
var kernelVersion int

func init() {
	str := C.CString("                                                                                                          ")
	size := 256

	C.sysctlbyname(C.CString("kern.osrelease"), unsafe.Pointer(str), (*C.size_t)(unsafe.Pointer(&size)), nil, 0)

	ver := C.GoString(str)
	kernelVersion, _ = strconv.Atoi(ver[:strings.Index(ver, ".")])

	mapping = make(map[int]version)

	mapping[16] = version{"10.12", "macOS", "Sierra"}
	mapping[15] = version{"10.11", "Mac OS X", "El Capitan"}
	mapping[14] = version{"10.10", "Mac OS X", "Yosemite"}
	mapping[13] = version{"10.9", "Mac OS X", "Mavericks"}
	mapping[12] = version{"10.8", "Mac OS X", "Mountain Lion"}
	mapping[11] = version{"10.7", "Mac OS X", "Lion"}
	mapping[10] = version{"10.6", "Mac OS X", "Snow Leopard"}
	mapping[9] = version{"10.5", "Mac OS X", "Leopard"}
	mapping[8] = version{"10.4", "Mac OS X", "Tiger"}
	mapping[7] = version{"10.3", "Mac OS X", "Panther"}
	mapping[6] = version{"10.2", "Mac OS X", "Jaguar"}
	mapping[5] = version{"10.1", "Mac OS X", "Puma"}
}

func GetVersion() string {
	_, ok := mapping[kernelVersion]

	if !ok {
		return ""
	}

	return mapping[kernelVersion].release
}

func GetDisplay() string {
	display := "macOS"

	release, ok := mapping[kernelVersion]

	if ok {
		display = release.name + " " + release.release + " " + release.releaseName
	}

	return display
}
