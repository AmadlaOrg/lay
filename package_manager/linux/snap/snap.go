package snap

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

// Client implements the package manager interface for snap
type Client struct{}

// Name returns the package manager name
func (s *Client) Name() string {
	return "snap"
}

// Install installs packages using snap install
func (s *Client) Install(packages []string) error {
	args := append([]string{"install"}, packages...)

	var cmd *exec.Cmd
	if osGetuid() != 0 {
		cmd = execCommand("sudo", append([]string{"snap"}, args...)...)
	} else {
		cmd = execCommand("snap", args...)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Search searches for packages using snap find
// Output table format:
// Name  Version  Publisher  Notes  Summary
func (s *Client) Search(query string) ([]types.SearchResult, error) {
	cmd := execCommand("snap", "find", query)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseSnapSearch(string(out)), nil
}

func parseSnapSearch(output string) []types.SearchResult {
	var results []types.SearchResult
	scanner := bufio.NewScanner(strings.NewReader(output))
	firstLine := true
	for scanner.Scan() {
		line := scanner.Text()
		// Skip header line
		if firstLine {
			firstLine = false
			continue
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		// Fields are whitespace-delimited: Name Version Publisher Notes Summary...
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		name := fields[0]
		version := fields[1]
		// Summary is everything from field 4 onward
		desc := strings.Join(fields[4:], " ")

		results = append(results, types.SearchResult{
			Name:        name,
			Version:     version,
			Description: desc,
		})
	}
	return results
}
