package jar

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectJava_Success(t *testing.T) {
	origExecCommand := execCommand
	defer func() { execCommand = origExecCommand }()

	execCommand = func(name string, args ...string) *exec.Cmd {
		// Return a command that prints java version info to combined output
		return exec.Command("echo", `openjdk version "21.0.1" 2024-01-16`)
	}

	info, err := DetectJava()
	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, "21.0.1", info.Version)
}

func TestDetectJava_Java8(t *testing.T) {
	origExecCommand := execCommand
	defer func() { execCommand = origExecCommand }()

	execCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("echo", `java version "1.8.0_392"`)
	}

	info, err := DetectJava()
	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, "1.8.0_392", info.Version)
}

func TestDetectJava_NotInstalled(t *testing.T) {
	origExecCommand := execCommand
	defer func() { execCommand = origExecCommand }()

	execCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("false") // exits with error
	}

	info, err := DetectJava()
	assert.Error(t, err)
	assert.Nil(t, info)
	assert.Contains(t, err.Error(), "java is not installed")
}

func TestDetectJava_UnparseableOutput(t *testing.T) {
	origExecCommand := execCommand
	defer func() { execCommand = origExecCommand }()

	execCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("echo", "some unexpected output")
	}

	info, err := DetectJava()
	assert.Error(t, err)
	assert.Nil(t, info)
	assert.Contains(t, err.Error(), "could not parse java version")
}

func TestIsJarFile(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"tika-app-3.2.3.jar", true},
		{"plantuml.JAR", true},
		{"tool.Jar", true},
		{"archive.tar.gz", false},
		{"binary", false},
		{"", false},
		{"something.jarx", false},
		{"file.jar.bak", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsJarFile(tt.filename))
		})
	}
}

func TestDeriveCommandName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"tika-app-3.2.3.jar", "tika-app"},
		{"plantuml-1.2024.8.jar", "plantuml"},
		{"closure-compiler-v20240317.jar", "closure-compiler"},
		{"lombok.jar", "lombok"},
		{"my-tool-2.0-beta.jar", "my-tool"},
		{"tool.jar", "tool"},
		{"some-lib-v1.0.jar", "some-lib"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, DeriveCommandName(tt.input))
		})
	}
}

func TestJarDir(t *testing.T) {
	origUserHomeDir := osUserHomeDir
	origMkdirAll := osMkdirAll
	defer func() {
		osUserHomeDir = origUserHomeDir
		osMkdirAll = origMkdirAll
	}()

	tmpDir := t.TempDir()
	osUserHomeDir = func() (string, error) { return tmpDir, nil }

	var createdDir string
	osMkdirAll = func(path string, perm os.FileMode) error {
		createdDir = path
		return nil
	}

	dir, err := JarDir()
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(tmpDir, ".local", "share", "lay", "jars"), dir)
	assert.Equal(t, dir, createdDir)
}

func TestJarDir_HomeError(t *testing.T) {
	origUserHomeDir := osUserHomeDir
	defer func() { osUserHomeDir = origUserHomeDir }()

	osUserHomeDir = func() (string, error) { return "", fmt.Errorf("no home") }

	_, err := JarDir()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not determine home directory")
}

func TestJarDir_MkdirError(t *testing.T) {
	origUserHomeDir := osUserHomeDir
	origMkdirAll := osMkdirAll
	defer func() {
		osUserHomeDir = origUserHomeDir
		osMkdirAll = origMkdirAll
	}()

	osUserHomeDir = func() (string, error) { return "/home/test", nil }
	osMkdirAll = func(path string, perm os.FileMode) error {
		return fmt.Errorf("permission denied")
	}

	_, err := JarDir()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not create JAR directory")
}

func TestCopyJar(t *testing.T) {
	origUserHomeDir := osUserHomeDir
	defer func() { osUserHomeDir = origUserHomeDir }()

	tmpDir := t.TempDir()
	osUserHomeDir = func() (string, error) { return tmpDir, nil }

	// Create a source JAR
	srcDir := t.TempDir()
	srcPath := filepath.Join(srcDir, "test-1.0.jar")
	err := os.WriteFile(srcPath, []byte("fake jar content"), 0644)
	assert.NoError(t, err)

	destPath, err := CopyJar(srcPath)
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(tmpDir, ".local", "share", "lay", "jars", "test-1.0.jar"), destPath)

	// Verify content was copied
	content, err := os.ReadFile(destPath)
	assert.NoError(t, err)
	assert.Equal(t, "fake jar content", string(content))
}

func TestCopyJar_SourceNotFound(t *testing.T) {
	origUserHomeDir := osUserHomeDir
	defer func() { osUserHomeDir = origUserHomeDir }()

	tmpDir := t.TempDir()
	osUserHomeDir = func() (string, error) { return tmpDir, nil }

	_, err := CopyJar("/nonexistent/file.jar")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not open source JAR")
}

func TestCreateLauncherScript_Unix(t *testing.T) {
	origGOOS := runtimeGOOS
	defer func() { runtimeGOOS = origGOOS }()
	runtimeGOOS = "linux"

	targetDir := t.TempDir()
	jarPath := "/home/user/.local/share/lay/jars/tika-app-3.2.3.jar"

	scriptPath, err := CreateLauncherScript(jarPath, "tika-app", targetDir)
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(targetDir, "tika-app"), scriptPath)

	content, err := os.ReadFile(scriptPath)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "#!/bin/sh")
	assert.Contains(t, string(content), "exec java -jar")
	assert.Contains(t, string(content), jarPath)
	assert.Contains(t, string(content), `"$@"`)

	info, err := os.Stat(scriptPath)
	assert.NoError(t, err)
	assert.True(t, info.Mode()&0111 != 0, "script should be executable")
}

func TestCreateLauncherScript_Windows(t *testing.T) {
	origGOOS := runtimeGOOS
	defer func() { runtimeGOOS = origGOOS }()
	runtimeGOOS = "windows"

	targetDir := t.TempDir()
	jarPath := `C:\Users\user\.local\share\lay\jars\tika-app-3.2.3.jar`

	scriptPath, err := CreateLauncherScript(jarPath, "tika-app", targetDir)
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(targetDir, "tika-app.cmd"), scriptPath)

	content, err := os.ReadFile(scriptPath)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "@echo off")
	assert.Contains(t, string(content), "java -jar")
	assert.Contains(t, string(content), jarPath)
	assert.Contains(t, string(content), "%*")
}

func TestCreateLauncherScript_WriteError(t *testing.T) {
	origGOOS := runtimeGOOS
	origWriteFile := osWriteFile
	defer func() {
		runtimeGOOS = origGOOS
		osWriteFile = origWriteFile
	}()
	runtimeGOOS = "linux"

	osWriteFile = func(name string, data []byte, perm os.FileMode) error {
		return fmt.Errorf("disk full")
	}

	_, err := CreateLauncherScript("/some/path.jar", "tool", "/target")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not write launcher script")
}

func TestParseLauncherScript(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		wantJar    string
		wantErr    bool
	}{
		{
			name:    "simple exec",
			content: "#!/bin/sh\nexec java -jar /path/to/tool.jar \"$@\"\n",
			wantJar: "/path/to/tool.jar",
		},
		{
			name:    "quoted path",
			content: "#!/bin/sh\nexec java -jar \"/path/to/tool.jar\" \"$@\"\n",
			wantJar: "/path/to/tool.jar",
		},
		{
			name:    "single-quoted path",
			content: "#!/bin/sh\nexec java -jar '/path/to/tool.jar' \"$@\"\n",
			wantJar: "/path/to/tool.jar",
		},
		{
			name:    "no exec prefix",
			content: "#!/bin/sh\njava -jar /path/to/tool.jar \"$@\"\n",
			wantJar: "/path/to/tool.jar",
		},
		{
			name:    "no java -jar invocation",
			content: "#!/bin/sh\necho hello\n",
			wantErr: true,
		},
		{
			name:    "empty file",
			content: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			scriptPath := filepath.Join(tmpDir, "launcher")
			err := os.WriteFile(scriptPath, []byte(tt.content), 0755)
			assert.NoError(t, err)

			jarPath, err := ParseLauncherScript(scriptPath)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantJar, jarPath)
			}
		})
	}
}

func TestParseLauncherScript_FileNotFound(t *testing.T) {
	_, err := ParseLauncherScript("/nonexistent/script")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "opening script")
}

func TestIsLauncherScript(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a launcher script
	launcherPath := filepath.Join(tmpDir, "launcher")
	err := os.WriteFile(launcherPath, []byte("#!/bin/sh\nexec java -jar /path/to/tool.jar \"$@\"\n"), 0755)
	assert.NoError(t, err)

	isLauncher, err := IsLauncherScript(launcherPath)
	assert.NoError(t, err)
	assert.True(t, isLauncher)

	// Create a non-launcher script
	normalPath := filepath.Join(tmpDir, "normal")
	err = os.WriteFile(normalPath, []byte("#!/bin/sh\necho hello\n"), 0755)
	assert.NoError(t, err)

	isLauncher, err = IsLauncherScript(normalPath)
	assert.NoError(t, err)
	assert.False(t, isLauncher)
}

func TestFindJarAssets(t *testing.T) {
	tests := []struct {
		name     string
		assets   []string
		expected []string
	}{
		{
			name:     "mixed assets with fat jar",
			assets:   []string{"tool.tar.gz", "tool-1.0.jar", "tool-1.0-app.jar", "tool-1.0-sources.jar"},
			expected: []string{"tool-1.0-app.jar", "tool-1.0.jar"},
		},
		{
			name:     "excludes sources and javadoc",
			assets:   []string{"lib-1.0-sources.jar", "lib-1.0-javadoc.jar", "lib-1.0-src.jar", "lib-1.0-doc.jar", "lib-1.0-tests.jar"},
			expected: []string{},
		},
		{
			name:     "no jar assets",
			assets:   []string{"tool-linux-amd64.tar.gz", "tool-darwin-arm64.tar.gz"},
			expected: []string{},
		},
		{
			name:     "standalone preferred",
			assets:   []string{"tool-1.0.jar", "tool-1.0-standalone.jar"},
			expected: []string{"tool-1.0-standalone.jar", "tool-1.0.jar"},
		},
		{
			name:     "all fat jar variants preferred",
			assets:   []string{"tool.jar", "tool-all.jar", "tool-fat.jar", "tool-uber.jar", "tool-complete.jar"},
			expected: []string{"tool-all.jar", "tool-fat.jar", "tool-uber.jar", "tool-complete.jar", "tool.jar"},
		},
		{
			name:     "single jar",
			assets:   []string{"plantuml-1.2024.8.jar"},
			expected: []string{"plantuml-1.2024.8.jar"},
		},
		{
			name:     "empty input",
			assets:   []string{},
			expected: []string{},
		},
		{
			name:     "only excluded jars means no results",
			assets:   []string{"lib-sources.jar", "lib-javadoc.jar"},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindJarAssets(tt.assets)
			assert.Equal(t, tt.expected, result)
		})
	}
}
