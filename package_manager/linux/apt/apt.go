package apt

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

// Client implements the package manager interface for apt
type Client struct{}

// Name returns the package manager name
func (s *Client) Name() string {
	return "apt"
}

// Install installs packages using apt
func (s *Client) Install(packages []string) error {
	args := append([]string{"install", "-y"}, packages...)

	var cmd *exec.Cmd
	if osGetuid() != 0 {
		cmd = execCommand("sudo", append([]string{"apt"}, args...)...)
	} else {
		cmd = execCommand("apt", args...)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Search searches for packages using apt-cache search
// Output format: "package - description"
func (s *Client) Search(query string) ([]types.SearchResult, error) {
	cmd := execCommand("apt-cache", "search", query)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseAptSearch(string(out)), nil
}

func parseAptSearch(output string) []types.SearchResult {
	var results []types.SearchResult
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// Format: "package - description"
		parts := strings.SplitN(line, " - ", 2)
		if len(parts) == 2 {
			results = append(results, types.SearchResult{
				Name:        strings.TrimSpace(parts[0]),
				Description: strings.TrimSpace(parts[1]),
			})
		}
	}
	return results
}
