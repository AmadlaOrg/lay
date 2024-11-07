package package_manager

import (
	"errors"
	"github.com/AmadlaOrg/lay/package_manager/linux"
	"github.com/AmadlaOrg/lay/package_manager/windows"
	"os"
	"runtime"
)

type IPackage interface{}

type SPackage struct{}

// FindPackageManager
func (s *SPackage) FindPackageManager(packageManagerName string) (string, error) {
	envPackageManager := os.Getenv("LAY_PACKAGE_MANAGER")
	if envPackageManager != "" {
		if runtime.GOOS == "windows" {
			if windows.Exists(envPackageManager) {
				return envPackageManager, nil
			}
		} else if runtime.GOOS == "linux" {
			if linux.Exists(envPackageManager) {
				return envPackageManager, nil
			}
		}
	} else if runtime.GOOS == "windows" {
		if windows.Exists(packageManagerName) {
			return envPackageManager, nil
		}
	} else if runtime.GOOS == "linux" {
		if linux.Exists(packageManagerName) {
			return envPackageManager, nil
		}
	}

	return "", errors.New("no package manager found")
}
