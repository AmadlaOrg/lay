package flatpak

// NewService creates a new flatpak package manager service
func NewService() *Client {
	return &Client{}
}
