package podman

// NewPodmanService to set up the apt service
func NewPodmanService() IPodman {
	return &SPodman{}
}
