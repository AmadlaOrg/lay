package choco

// NewService creates a new choco package manager service
func NewService() Manager {
	return &Client{}
}
