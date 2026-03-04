package package_manager

import (
	"fmt"

	"github.com/AmadlaOrg/lay/package_manager/linux/apk"
	"github.com/AmadlaOrg/lay/package_manager/linux/apt"
	"github.com/AmadlaOrg/lay/package_manager/linux/dnf"
	"github.com/AmadlaOrg/lay/package_manager/linux/dpkg"
	"github.com/AmadlaOrg/lay/package_manager/linux/flatpak"
	"github.com/AmadlaOrg/lay/package_manager/linux/nix"
	"github.com/AmadlaOrg/lay/package_manager/linux/pacman"
	"github.com/AmadlaOrg/lay/package_manager/linux/rpm"
	"github.com/AmadlaOrg/lay/package_manager/linux/snap"
	"github.com/AmadlaOrg/lay/package_manager/linux/yum"
	"github.com/AmadlaOrg/lay/package_manager/linux/zypper"
	"github.com/AmadlaOrg/lay/package_manager/macos/brew"
	"github.com/AmadlaOrg/lay/package_manager/windows/choco"
	"github.com/AmadlaOrg/lay/package_manager/windows/scoop"
	"github.com/AmadlaOrg/lay/package_manager/windows/winget"
)

// NewManagerByName returns a Manager implementation for the given name
func NewManagerByName(name string) (Manager, error) {
	switch name {
	case "apt":
		return apt.NewService(), nil
	case "dnf":
		return dnf.NewService(), nil
	case "yum":
		return yum.NewService(), nil
	case "pacman":
		return pacman.NewService(), nil
	case "zypper":
		return zypper.NewService(), nil
	case "apk":
		return apk.NewService(), nil
	case "nix":
		return nix.NewService(), nil
	case "snap":
		return snap.NewService(), nil
	case "flatpak":
		return flatpak.NewService(), nil
	case "dpkg":
		return dpkg.NewService(), nil
	case "rpm":
		return rpm.NewService(), nil
	case "choco":
		return choco.NewService(), nil
	case "scoop":
		return scoop.NewService(), nil
	case "winget":
		return winget.NewService(), nil
	case "brew":
		return brew.NewService(), nil
	default:
		return nil, fmt.Errorf("unsupported package manager: %s", name)
	}
}
