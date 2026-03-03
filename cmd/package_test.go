package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPackageCmd_ManagerFlagRegistered(t *testing.T) {
	flag := PackageCmd.PersistentFlags().Lookup("manager")
	assert.NotNil(t, flag)
	assert.Equal(t, "", flag.DefValue)
}

func TestPackageCmd_HasInstallSubcommand(t *testing.T) {
	found := false
	for _, sub := range PackageCmd.Commands() {
		if sub.Name() == "install" {
			found = true
			break
		}
	}
	assert.True(t, found, "install subcommand should be registered")
}

func TestPackageCmd_HasSearchSubcommand(t *testing.T) {
	found := false
	for _, sub := range PackageCmd.Commands() {
		if sub.Name() == "search" {
			found = true
			break
		}
	}
	assert.True(t, found, "search subcommand should be registered")
}
