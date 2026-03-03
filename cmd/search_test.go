package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchCmd_RequiresExactlyOneArg(t *testing.T) {
	err := searchCmd.Args(searchCmd, []string{})
	assert.Error(t, err)
}

func TestSearchCmd_AcceptsOneArg(t *testing.T) {
	err := searchCmd.Args(searchCmd, []string{"curl"})
	assert.NoError(t, err)
}

func TestSearchCmd_RejectsTwoArgs(t *testing.T) {
	err := searchCmd.Args(searchCmd, []string{"curl", "wget"})
	assert.Error(t, err)
}
