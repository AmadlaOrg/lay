package docker

// NewService creates a new docker container runtime service
func NewService() *Client {
	return &Client{}
}
