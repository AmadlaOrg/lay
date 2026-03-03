package cargo

// NewService creates a new cargo build system service
func NewService() *Builder {
	return &Builder{}
}
