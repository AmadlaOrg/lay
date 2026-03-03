package dpkg

// NewService creates a new dpkg package manager service
func NewService() *Client {
	return &Client{}
}
