package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainerCmd_DisablesFlagParsing(t *testing.T) {
	assert.True(t, ContainerCmd.DisableFlagParsing)
}

func TestExtractRuntimeFlag_WithSeparateValue(t *testing.T) {
	runtime, remaining := extractRuntimeFlag([]string{"--runtime", "docker", "ps"})
	assert.Equal(t, "docker", runtime)
	assert.Equal(t, []string{"ps"}, remaining)
}

func TestExtractRuntimeFlag_WithEqualsValue(t *testing.T) {
	runtime, remaining := extractRuntimeFlag([]string{"--runtime=podman", "ps"})
	assert.Equal(t, "podman", runtime)
	assert.Equal(t, []string{"ps"}, remaining)
}

func TestExtractRuntimeFlag_NoFlag(t *testing.T) {
	runtime, remaining := extractRuntimeFlag([]string{"run", "--rm", "alpine"})
	assert.Equal(t, "", runtime)
	assert.Equal(t, []string{"run", "--rm", "alpine"}, remaining)
}

func TestExtractRuntimeFlag_Empty(t *testing.T) {
	runtime, remaining := extractRuntimeFlag([]string{})
	assert.Equal(t, "", runtime)
	assert.Nil(t, remaining)
}

func TestExtractRuntimeFlag_MixedWithContainerFlags(t *testing.T) {
	runtime, remaining := extractRuntimeFlag([]string{"--runtime", "docker", "run", "--rm", "-it", "alpine"})
	assert.Equal(t, "docker", runtime)
	assert.Equal(t, []string{"run", "--rm", "-it", "alpine"}, remaining)
}
