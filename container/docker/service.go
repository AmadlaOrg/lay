package docker

// NewDockerService to set up the docker service
func NewDockerService() IDocker {
	return &SDocker{}
}
