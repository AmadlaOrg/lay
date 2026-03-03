package linux

import "fmt"

type Values int

// Enum values
const (
	aptVal Values = iota
	dnfVal
	dpkgVal
	rpmVal
	yumVal
	pacmanVal
	zypperVal
	apkVal
	nixVal
	snapVal
	flatpakVal
)

// String method to convert enum values to their string representation
func (r Values) String() string {
	values := [...]string{
		"apt",
		"dnf",
		"dpkg",
		"rpm",
		"yum",
		"pacman",
		"zypper",
		"apk",
		"nix",
		"snap",
		"flatpak",
	}

	if int(r) < 0 || int(r) >= len(values) {
		return fmt.Sprintf("Unknown(%d)", r) // Handle invalid enum values
	}
	return values[r]
}

// Exists checks if a given package manager name exist
func Exists(name string) bool {
	for _, v := range [...]Values{aptVal, dnfVal, dpkgVal, rpmVal, yumVal, pacmanVal, zypperVal, apkVal, nixVal, snapVal, flatpakVal} {
		if v.String() == name {
			return true
		}
	}
	return false
}
