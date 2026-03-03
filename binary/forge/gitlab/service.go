package gitlab

import "github.com/AmadlaOrg/lay/binary/forge"

// NewService creates a new GitLab forge service
func NewService() forge.Forge {
	return &Client{}
}
