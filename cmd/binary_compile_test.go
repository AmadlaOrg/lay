package cmd

import (
	"os"
	"path/filepath"
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

func TestCopyBinaryFile_Success(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	srcPath := filepath.Join(srcDir, "mybinary")
	dstPath := filepath.Join(dstDir, "mybinary")
	content := []byte("#!/bin/sh\necho hello")

	err := os.WriteFile(srcPath, content, 0755)
	assert.NoError(t, err)

	err = copyBinaryFile(srcPath, dstPath)
	assert.NoError(t, err)

	got, err := os.ReadFile(dstPath)
	assert.NoError(t, err)
	assert.Equal(t, content, got)

	info, err := os.Stat(dstPath)
	assert.NoError(t, err)
	assert.True(t, info.Mode()&0111 != 0, "binary should be executable")
}

func TestCopyBinaryFile_SourceMissing(t *testing.T) {
	dstDir := t.TempDir()
	dstPath := filepath.Join(dstDir, "mybinary")

	err := copyBinaryFile("/nonexistent/path/binary", dstPath)
	assert.Error(t, err)
}
