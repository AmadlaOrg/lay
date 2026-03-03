package rpm

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

// Client implements the package manager interface for rpm (local .rpm file installer)
type Client struct{}

// Name returns the package manager name
func (s *Client) Name() string {
	return "rpm"
}

// Install installs local .rpm files using rpm -i
func (s *Client) Install(packages []string) error {
	args := append([]string{"-i"}, packages...)

	var cmd *exec.Cmd
	if osGetuid() != 0 {
		cmd = execCommand("sudo", append([]string{"rpm"}, args...)...)
	} else {
		cmd = execCommand("rpm", args...)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Search lists installed packages matching the query using rpm -qa
// Output: tab-separated name, version, summary per line
func (s *Client) Search(query string) ([]types.SearchResult, error) {
	cmd := execCommand("rpm", "-qa", "--queryformat", "%{NAME}\t%{VERSION}-%{RELEASE}\t%{SUMMARY}\n", "*"+query+"*")
	out, err := cmd.Output()
	if err != nil {
		return nil, nil //nolint:nilerr // intentional: non-zero exit means no results
	}
	return parseRpmSearch(string(out)), nil
}

func parseRpmSearch(output string) []types.SearchResult {
	var results []types.SearchResult
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		cols := strings.SplitN(line, "\t", 3)
		if len(cols) >= 3 {
			results = append(results, types.SearchResult{
				Name:        cols[0],
				Version:     cols[1],
				Description: cols[2],
			})
		}
	}
	return results
}
