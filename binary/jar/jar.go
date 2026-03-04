package jar

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
)

// For testability
var (
	execCommand   = exec.Command
	osUserHomeDir = os.UserHomeDir
	osMkdirAll    = os.MkdirAll
	osWriteFile   = os.WriteFile
	osOpen        = os.Open
	runtimeGOOS   = runtime.GOOS
)

// JavaInfo holds information about the detected Java installation
type JavaInfo struct {
	Version string // e.g. "21.0.1"
	Path    string // path to java binary
}

// DetectJava runs `java -version` and parses the output.
// Returns error if java is not installed.
func DetectJava() (*JavaInfo, error) {
	cmd := execCommand("java", "-version")
	// java -version writes to stderr
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("java is not installed or not in PATH: %w", err)
	}

	version := parseJavaVersion(string(output))
	if version == "" {
		return nil, fmt.Errorf("could not parse java version from output: %s", string(output))
	}

	javaPath, _ := exec.LookPath("java")

	return &JavaInfo{
		Version: version,
		Path:    javaPath,
	}, nil
}

// parseJavaVersion extracts the version string from `java -version` output.
// Handles formats like:
//
//	openjdk version "21.0.1" ...
//	java version "1.8.0_392" ...
var javaVersionRe = regexp.MustCompile(`(?:java|openjdk) version "([^"]+)"`)

func parseJavaVersion(output string) string {
	matches := javaVersionRe.FindStringSubmatch(output)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

// IsJarFile returns true if filename ends with .jar (case-insensitive)
func IsJarFile(filename string) bool {
	return strings.HasSuffix(strings.ToLower(filename), ".jar")
}

// DeriveCommandName extracts a command name from a JAR filename.
// Strips .jar suffix and version patterns.
//
// Examples:
//
//	tika-app-3.2.3.jar      → tika-app
//	plantuml-1.2024.8.jar   → plantuml
//	closure-compiler-v20240317.jar → closure-compiler
//	lombok.jar              → lombok
var versionSegmentRe = regexp.MustCompile(`^v?\d`)

func DeriveCommandName(jarFilename string) string {
	// Strip .jar suffix (case-insensitive)
	name := jarFilename
	if strings.HasSuffix(strings.ToLower(name), ".jar") {
		name = name[:len(name)-4]
	}

	// Split by '-' and find first version segment
	parts := strings.Split(name, "-")
	for i, part := range parts {
		if versionSegmentRe.MatchString(part) {
			if i > 0 {
				return strings.Join(parts[:i], "-")
			}
			// Version at position 0 — use full stem
			return name
		}
	}

	return name
}

// JarDir returns ~/.local/share/lay/jars/, creating it if needed.
func JarDir() (string, error) {
	home, err := osUserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}

	dir := filepath.Join(home, ".local", "share", "lay", "jars")
	if err := osMkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("could not create JAR directory: %w", err)
	}

	return dir, nil
}

// CopyJar copies a JAR file to the jars directory, returns the destination path.
func CopyJar(srcPath string) (string, error) {
	jarDir, err := JarDir()
	if err != nil {
		return "", err
	}

	filename := filepath.Base(srcPath)
	destPath := filepath.Join(jarDir, filename)

	src, err := os.Open(srcPath)
	if err != nil {
		return "", fmt.Errorf("could not open source JAR: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("could not create destination JAR: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("could not copy JAR: %w", err)
	}

	return destPath, nil
}

// CreateLauncherScript creates a wrapper script in targetDir.
// On Unix: shell script. On Windows: .cmd batch file.
// Returns the path to the created script.
func CreateLauncherScript(jarPath, commandName, targetDir string) (string, error) {
	if runtimeGOOS == "windows" {
		return createWindowsLauncher(jarPath, commandName, targetDir)
	}
	return createUnixLauncher(jarPath, commandName, targetDir)
}

func createUnixLauncher(jarPath, commandName, targetDir string) (string, error) {
	scriptPath := filepath.Join(targetDir, commandName)

	content := fmt.Sprintf("#!/bin/sh\nexec java -jar %q \"$@\"\n", jarPath)

	if err := osWriteFile(scriptPath, []byte(content), 0755); err != nil {
		return "", fmt.Errorf("could not write launcher script: %w", err)
	}

	return scriptPath, nil
}

func createWindowsLauncher(jarPath, commandName, targetDir string) (string, error) {
	scriptPath := filepath.Join(targetDir, commandName+".cmd")

	content := fmt.Sprintf("@echo off\r\njava -jar \"%s\" %%*\r\n", jarPath)

	if err := osWriteFile(scriptPath, []byte(content), 0755); err != nil {
		return "", fmt.Errorf("could not write launcher script: %w", err)
	}

	return scriptPath, nil
}

// FindJarAssets filters asset names for installable JAR files.
// Excludes sources/javadoc/tests JARs. Prefers fat/app/standalone JARs.
// Returns asset names sorted by preference (best first).
func FindJarAssets(assetNames []string) []string {
	var results []scoredAsset

	for _, name := range assetNames {
		if !IsJarFile(name) {
			continue
		}

		lower := strings.ToLower(name)

		// Exclude non-runnable JARs
		if isExcludedJar(lower) {
			continue
		}

		score := 0
		// Prefer fat/standalone/app JARs (these are runnable)
		if isFatJar(lower) {
			score += 10
		}

		results = append(results, scoredAsset{name: name, score: score})
	}

	// Sort descending by score
	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	names := make([]string, len(results))
	for i, r := range results {
		names[i] = r.name
	}
	return names
}

type scoredAsset struct {
	name  string
	score int
}

var excludedSuffixes = []string{
	"-sources.jar",
	"-javadoc.jar",
	"-src.jar",
	"-doc.jar",
	"-tests.jar",
}

func isExcludedJar(lowerName string) bool {
	for _, suffix := range excludedSuffixes {
		if strings.HasSuffix(lowerName, suffix) {
			return true
		}
	}
	return false
}

var fatJarPatterns = []string{
	"-app",
	"-standalone",
	"-all",
	"-fat",
	"-uber",
	"-complete",
}

func isFatJar(lowerName string) bool {
	for _, pattern := range fatJarPatterns {
		if strings.Contains(lowerName, pattern) {
			return true
		}
	}
	return false
}

// javaJarRe matches java -jar invocations in launcher scripts.
// Handles optional "exec", optional quotes around the jar path.
var javaJarRe = regexp.MustCompile(`(?:exec\s+)?java\s+-jar\s+["']?([^"'\s]+)["']?`)

// ParseLauncherScript reads a shell script and extracts the JAR path
// from a `java -jar <path>` invocation.
func ParseLauncherScript(scriptPath string) (string, error) {
	f, err := osOpen(scriptPath)
	if err != nil {
		return "", fmt.Errorf("opening script: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if matches := javaJarRe.FindStringSubmatch(line); len(matches) >= 2 {
			return matches[1], nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("reading script: %w", err)
	}
	return "", fmt.Errorf("no java -jar invocation found in %s", scriptPath)
}

// IsLauncherScript returns true if the file at path is a shell script
// containing a `java -jar` invocation.
func IsLauncherScript(path string) (bool, error) {
	_, err := ParseLauncherScript(path)
	if err != nil {
		return false, nil
	}
	return true, nil
}
