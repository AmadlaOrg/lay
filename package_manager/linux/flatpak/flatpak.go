package flatpak

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

// Client implements the package manager interface for flatpak
type Client struct{}

// Name returns the package manager name
func (s *Client) Name() string {
	return "flatpak"
}

// Install installs packages using flatpak install
func (s *Client) Install(packages []string) error {
	args := append([]string{"install", "-y"}, packages...)

	var cmd *exec.Cmd
	if osGetuid() != 0 {
		cmd = execCommand("sudo", append([]string{"flatpak"}, args...)...)
	} else {
		cmd = execCommand("flatpak", args...)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Search searches for packages using flatpak search
// Output columns: Name\tDescription\tApplication ID\tVersion\tBranch\tRemotes
func (s *Client) Search(query string) ([]types.SearchResult, error) {
	cmd := execCommand("flatpak", "search", query)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseFlatpakSearch(string(out)), nil
}

func parseFlatpakSearch(output string) []types.SearchResult {
	var results []types.SearchResult
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		// Flatpak uses tab-separated columns:
		// Name\tDescription\tApplication ID\tVersion\tBranch\tRemotes
		cols := strings.Split(line, "\t")
		if len(cols) >= 4 {
			results = append(results, types.SearchResult{
				Name:        strings.TrimSpace(cols[0]),
				Version:     strings.TrimSpace(cols[3]),
				Description: strings.TrimSpace(cols[1]),
			})
		}
	}
	return results
}
