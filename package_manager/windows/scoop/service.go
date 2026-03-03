package scoop

// NewService creates a new scoop package manager service
func NewService() Manager {
	return &Client{}
}
