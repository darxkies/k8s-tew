package oslib

import (
	"fmt"
	"testing"
)

func Test(t *testing.T) {
	fmt.Println("Name:", Name)
	fmt.Println("Version:", GetVersion())
	fmt.Println("Display:", GetDisplay())
	fmt.Println("Arch:", GetDisplayArch())

	if Name == "Linux" {
		dist := GetDist()
		fmt.Println("Dist:", dist.Display)
		fmt.Println("Release:", dist.Release)
		fmt.Println("Codename:", dist.Codename)
	}
}
