package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBinaryInstallCmd_RequiresExactlyOneArg(t *testing.T) {
	err := binaryInstallCmd.Args(binaryInstallCmd, []string{})
	assert.Error(t, err)
}

func TestBinaryInstallCmd_AcceptsOneArg(t *testing.T) {
	err := binaryInstallCmd.Args(binaryInstallCmd, []string{"sharkdp/fd"})
	assert.NoError(t, err)
}

func TestBinaryInstallCmd_RejectsTwoArgs(t *testing.T) {
	err := binaryInstallCmd.Args(binaryInstallCmd, []string{"sharkdp/fd", "extra"})
	assert.Error(t, err)
}
