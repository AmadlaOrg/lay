package podman

// NewService creates a new podman container runtime service
func NewService() *Client {
	return &Client{}
}
