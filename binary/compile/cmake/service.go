package cmake

// NewService creates a new cmake build system service
func NewService() *Builder {
	return &Builder{}
}
