package podman

// NewPodmanService to set up the podman service
func NewPodmanService() IPodman {
	return &SPodman{}
}
