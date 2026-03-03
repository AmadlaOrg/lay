package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInstallCmd_RequiresArgs(t *testing.T) {
	err := installCmd.Args(installCmd, []string{})
	assert.Error(t, err)
}

func TestInstallCmd_AcceptsArgs(t *testing.T) {
	err := installCmd.Args(installCmd, []string{"curl"})
	assert.NoError(t, err)
}

func TestInstallCmd_AcceptsMultipleArgs(t *testing.T) {
	err := installCmd.Args(installCmd, []string{"curl", "wget", "vim"})
	assert.NoError(t, err)
}
