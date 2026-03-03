package zypper

// NewService creates a new zypper package manager service
func NewService() *Client {
	return &Client{}
}
