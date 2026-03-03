package makefile

// NewService creates a new makefile build system service
func NewService() *Builder {
	return &Builder{}
}
