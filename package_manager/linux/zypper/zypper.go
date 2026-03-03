package zypper

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

// Client implements the package manager interface for zypper
type Client struct{}

// Name returns the package manager name
func (s *Client) Name() string {
	return "zypper"
}

// Install installs packages using zypper
func (s *Client) Install(packages []string) error {
	args := append([]string{"install", "-y"}, packages...)

	var cmd *exec.Cmd
	if osGetuid() != 0 {
		cmd = execCommand("sudo", append([]string{"zypper"}, args...)...)
	} else {
		cmd = execCommand("zypper", args...)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Search searches for packages using zypper search
// Output table format with | delimiters:
// S | Name | Summary | Type
func (s *Client) Search(query string) ([]types.SearchResult, error) {
	cmd := execCommand("zypper", "search", query)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseZypperSearch(string(out)), nil
}

func parseZypperSearch(output string) []types.SearchResult {
	var results []types.SearchResult
	scanner := bufio.NewScanner(strings.NewReader(output))
	headerPassed := false
	for scanner.Scan() {
		line := scanner.Text()
		// Skip until we pass the header separator (--+--)
		if strings.Contains(line, "--+--") {
			headerPassed = true
			continue
		}
		if !headerPassed || strings.TrimSpace(line) == "" {
			continue
		}
		// Format: "S | name | summary | type"
		cols := strings.Split(line, "|")
		if len(cols) >= 3 {
			results = append(results, types.SearchResult{
				Name:        strings.TrimSpace(cols[1]),
				Description: strings.TrimSpace(cols[2]),
			})
		}
	}
	return results
}
