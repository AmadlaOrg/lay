package target

// NewService creates a new target service
func NewService() Target {
	return &Service{}
}
