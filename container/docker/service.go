package docker

// NewDockerService to set up the apt service
func NewDockerService() IDocker {
	return &SDocker{}
}
