package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBinaryCmd_HasSubcommands(t *testing.T) {
	cmds := BinaryCmd.Commands()
	names := make([]string, len(cmds))
	for i, c := range cmds {
		names[i] = c.Name()
	}
	assert.Contains(t, names, "install")
	assert.Contains(t, names, "compile")
}

func TestBinaryCmd_HasToFlag(t *testing.T) {
	flag := BinaryCmd.PersistentFlags().Lookup("to")
	assert.NotNil(t, flag)
}
