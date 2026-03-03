package github

import "github.com/AmadlaOrg/lay/binary/forge"

// NewService creates a new GitHub forge service
func NewService() forge.Forge {
	return &Client{}
}
