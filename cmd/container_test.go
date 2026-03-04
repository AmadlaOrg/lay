package cmd

import (
	"errors"
	"testing"

	ct "github.com/AmadlaOrg/lay/container"
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

type mockContainerDetector struct {
	name string
	err  error
}

func (m *mockContainerDetector) Detect(override string) (string, error) {
	return m.name, m.err
}

type mockRuntime struct {
	name    string
	execErr error
}

func (m *mockRuntime) Name() string            { return m.name }
func (m *mockRuntime) Exec(args []string) error { return m.execErr }

func TestRunExecWithRuntime_Success(t *testing.T) {
	detector := &mockContainerDetector{name: "docker"}
	rt := &mockRuntime{name: "docker"}
	runtimeFn := func(name string) (ct.Runtime, error) { return rt, nil }

	err := runExecWithRuntime(detector, runtimeFn, "", []string{"ps"}, newTestOut())
	assert.NoError(t, err)
}

func TestRunExecWithRuntime_DetectError(t *testing.T) {
	detector := &mockContainerDetector{err: errors.New("no runtime")}
	runtimeFn := func(name string) (ct.Runtime, error) { return nil, nil }

	err := runExecWithRuntime(detector, runtimeFn, "", []string{"ps"}, newTestOut())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no runtime")
}

func TestRunExecWithRuntime_RuntimeError(t *testing.T) {
	detector := &mockContainerDetector{name: "bad"}
	runtimeFn := func(name string) (ct.Runtime, error) { return nil, errors.New("unsupported runtime") }

	err := runExecWithRuntime(detector, runtimeFn, "", []string{"ps"}, newTestOut())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported runtime")
}

func TestRunExecWithRuntime_ExecError(t *testing.T) {
	detector := &mockContainerDetector{name: "docker"}
	rt := &mockRuntime{name: "docker", execErr: errors.New("exec failed")}
	runtimeFn := func(name string) (ct.Runtime, error) { return rt, nil }

	err := runExecWithRuntime(detector, runtimeFn, "", []string{"bad-cmd"}, newTestOut())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exec failed")
}

func TestExecWithRuntime_Error(t *testing.T) {
	origDetector := newContainerDetector
	origExit := osExit
	defer func() {
		newContainerDetector = origDetector
		osExit = origExit
	}()

	newContainerDetector = func() ct.Detector {
		return &mockContainerDetector{err: errors.New("no runtime")}
	}

	var exitCode int
	osExit = func(code int) { exitCode = code }

	execWithRuntime("", []string{"ps"})
	assert.Equal(t, 1, exitCode)
}

func TestExecWithRuntime_Success(t *testing.T) {
	origDetector := newContainerDetector
	origRuntime := newRuntimeByName
	origExit := osExit
	defer func() {
		newContainerDetector = origDetector
		newRuntimeByName = origRuntime
		osExit = origExit
	}()

	newContainerDetector = func() ct.Detector {
		return &mockContainerDetector{name: "docker"}
	}
	newRuntimeByName = func(name string) (ct.Runtime, error) {
		return &mockRuntime{name: "docker"}, nil
	}

	var exitCode int = -1
	osExit = func(code int) { exitCode = code }

	execWithRuntime("", []string{"ps"})
	assert.Equal(t, -1, exitCode, "should not call osExit on success")
}
