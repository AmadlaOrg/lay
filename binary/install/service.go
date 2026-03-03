package install

// NewInstallService creates a new install service
func NewInstallService() Installer {
	return &Service{}
}
