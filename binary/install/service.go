package install

// New creates a new Installer.
func New() Installer {
	return &Service{}
}

// NewWithName creates a new Installer with a command name override.
func NewWithName(name string) Installer {
	return &Service{NameOverride: name}
}
