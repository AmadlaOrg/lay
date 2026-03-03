package dnf

// NewService creates a new dnf package manager service
func NewService() *Client {
	return &Client{}
}
