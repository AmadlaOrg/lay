package brew

// NewService creates a new brew package manager service
func NewService() *Client {
	return &Client{}
}
