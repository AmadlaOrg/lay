package winget

// NewService creates a new winget package manager service
func NewService() Manager {
	return &Client{}
}
