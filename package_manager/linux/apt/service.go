package apt

// NewService creates a new apt package manager service
func NewService() *Client {
	return &Client{}
}
