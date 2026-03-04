package install

// NewInstallService creates a new install service
func NewInstallService() Installer {
	return &Service{}
}

// NewInstallServiceWithName creates a new install service with a command name override
func NewInstallServiceWithName(name string) Installer {
	return &Service{NameOverride: name}
}
