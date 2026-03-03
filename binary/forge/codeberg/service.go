package codeberg

import "github.com/AmadlaOrg/lay/binary/forge"

// NewService creates a new Codeberg forge service
func NewService() forge.Forge {
	return &Client{}
}
