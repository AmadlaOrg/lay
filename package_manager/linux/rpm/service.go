package rpm

// NewService creates a new rpm package manager service
func NewService() *Client {
	return &Client{}
}
