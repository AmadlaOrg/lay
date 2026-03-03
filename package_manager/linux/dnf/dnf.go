package dnf

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

// Client implements the package manager interface for dnf
type Client struct{}

// Name returns the package manager name
func (s *Client) Name() string {
	return "dnf"
}

// Install installs packages using dnf
func (s *Client) Install(packages []string) error {
	args := append([]string{"install", "-y"}, packages...)

	var cmd *exec.Cmd
	if osGetuid() != 0 {
		cmd = execCommand("sudo", append([]string{"dnf"}, args...)...)
	} else {
		cmd = execCommand("dnf", args...)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Search searches for packages using dnf search
// Output format lines like: "name.arch : description"
func (s *Client) Search(query string) ([]types.SearchResult, error) {
	cmd := execCommand("dnf", "search", "-q", query)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseDnfSearch(string(out)), nil
}

func parseDnfSearch(output string) []types.SearchResult {
	var results []types.SearchResult
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "=") {
			continue
		}
		// Format: "name.arch : description"
		parts := strings.SplitN(line, " : ", 2)
		if len(parts) == 2 {
			nameArch := strings.TrimSpace(parts[0])
			// Strip .arch suffix (e.g., "curl.x86_64" → "curl")
			name := nameArch
			if idx := strings.LastIndex(nameArch, "."); idx > 0 {
				name = nameArch[:idx]
			}
			results = append(results, types.SearchResult{
				Name:        name,
				Description: strings.TrimSpace(parts[1]),
			})
		}
	}
	return results
}
