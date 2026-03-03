package nix

// NewService creates a new nix package manager service
func NewService() *Client {
	return &Client{}
}
