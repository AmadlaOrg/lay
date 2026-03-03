package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBinaryCompileCmd_RequiresExactlyOneArg(t *testing.T) {
	err := binaryCompileCmd.Args(binaryCompileCmd, []string{})
	assert.Error(t, err)
}

func TestBinaryCompileCmd_AcceptsOneArg(t *testing.T) {
	err := binaryCompileCmd.Args(binaryCompileCmd, []string{"."})
	assert.NoError(t, err)
}

func TestBinaryCompileCmd_HasBuildSystemFlag(t *testing.T) {
	flag := binaryCompileCmd.Flags().Lookup("build-system")
	assert.NotNil(t, flag)
}
