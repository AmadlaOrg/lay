package yum

// NewService creates a new yum package manager service
func NewService() *Client {
	return &Client{}
}
