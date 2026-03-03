package apk

import (
	"bufio"
	"os"
	"os/exec"
	"strings"

	"github.com/AmadlaOrg/lay/package_manager/types"
)

// For mocking
var (
	execCommand = exec.Command
	osGetuid    = os.Getuid
)

// Client implements the package manager interface for apk (Alpine)
type Client struct{}

// Name returns the package manager name
func (s *Client) Name() string {
	return "apk"
}

// Install installs packages using apk add
func (s *Client) Install(packages []string) error {
	args := append([]string{"add"}, packages...)

	var cmd *exec.Cmd
	if osGetuid() != 0 {
		cmd = execCommand("sudo", append([]string{"apk"}, args...)...)
	} else {
		cmd = execCommand("apk", args...)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Search searches for packages using apk search -v
// Output format: "package-version - description"
func (s *Client) Search(query string) ([]types.SearchResult, error) {
	cmd := execCommand("apk", "search", "-v", query)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseApkSearch(string(out)), nil
}

func parseApkSearch(output string) []types.SearchResult {
	var results []types.SearchResult
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// Format: "package-version - description"
		parts := strings.SplitN(line, " - ", 2)
		if len(parts) == 2 {
			nameVersion := strings.TrimSpace(parts[0])
			// Split name from version: last hyphen before a digit
			name, version := splitApkNameVersion(nameVersion)
			results = append(results, types.SearchResult{
				Name:        name,
				Version:     version,
				Description: strings.TrimSpace(parts[1]),
			})
		}
	}
	return results
}

// splitApkNameVersion splits "curl-8.5.0-r0" into ("curl", "8.5.0-r0")
func splitApkNameVersion(nameVersion string) (string, string) {
	// Find the last '-' that precedes a digit
	for i := len(nameVersion) - 1; i >= 0; i-- {
		if nameVersion[i] == '-' && i+1 < len(nameVersion) && nameVersion[i+1] >= '0' && nameVersion[i+1] <= '9' {
			return nameVersion[:i], nameVersion[i+1:]
		}
	}
	return nameVersion, ""
}
