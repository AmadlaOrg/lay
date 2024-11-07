package package_manager

// NewPackageManagerService to set up the package manager service
func NewPackageManagerService() IPackageManager {
	return &SPackageManager{}
}
