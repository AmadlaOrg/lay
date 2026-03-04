package rpm

import (
	"bufio"
	"errors"
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

// Remove is not supported by rpm (use dnf or yum instead)
func (s *Client) Remove(_ []string) error {
	return errors.New("rpm does not support package removal — use dnf or yum instead")
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

// Update is not supported by rpm (it is a local installer, not a repo manager)
func (s *Client) Update() error {
	return types.ErrNotSupported
}

// Upgrade is not supported by rpm (it is a local installer, not a repo manager)
func (s *Client) Upgrade(packages []string) error {
	return types.ErrNotSupported
}

// List returns all installed packages using rpm -qa
func (s *Client) List() ([]types.PackageInfo, error) {
	cmd := execCommand("rpm", "-qa", "--queryformat", "%{NAME}\t%{VERSION}-%{RELEASE}\n")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseRpmList(string(out)), nil
}

// IsInstalled checks whether a package is installed using rpm -q
func (s *Client) IsInstalled(pkg string) (bool, error) {
	cmd := execCommand("rpm", "-q", pkg)
	err := cmd.Run()
	if err != nil {
		return false, nil
	}
	return true, nil
}

func parseRpmList(output string) []types.PackageInfo {
	var results []types.PackageInfo
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		cols := strings.SplitN(line, "\t", 2)
		if len(cols) >= 2 {
			results = append(results, types.PackageInfo{
				Name:    cols[0],
				Version: cols[1],
			})
		}
	}
	return results
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
