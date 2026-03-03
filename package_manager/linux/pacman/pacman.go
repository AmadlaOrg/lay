package pacman

import (
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

// Client implements the package manager interface for pacman
type Client struct{}

// Name returns the package manager name
func (s *Client) Name() string {
	return "pacman"
}

// Install installs packages using pacman
func (s *Client) Install(packages []string) error {
	args := append([]string{"-S", "--noconfirm"}, packages...)

	var cmd *exec.Cmd
	if osGetuid() != 0 {
		cmd = execCommand("sudo", append([]string{"pacman"}, args...)...)
	} else {
		cmd = execCommand("pacman", args...)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Search searches for packages using pacman -Ss
// Output format:
//
//	repo/package version
//	    description
func (s *Client) Search(query string) ([]types.SearchResult, error) {
	cmd := execCommand("pacman", "-Ss", query)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parsePacmanSearch(string(out)), nil
}

func parsePacmanSearch(output string) []types.SearchResult {
	var results []types.SearchResult
	lines := strings.Split(output, "\n")
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if line == "" || strings.HasPrefix(line, " ") {
			continue
		}
		// Header line: "repo/name version [installed]"
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		repoName := parts[0]
		version := parts[1]

		name := repoName
		if idx := strings.Index(repoName, "/"); idx >= 0 {
			name = repoName[idx+1:]
		}

		var desc string
		if i+1 < len(lines) {
			desc = strings.TrimSpace(lines[i+1])
			i++ // skip description line
		}

		results = append(results, types.SearchResult{
			Name:        name,
			Version:     version,
			Description: desc,
		})
	}
	return results
}
