package dpkg

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

// Client implements the package manager interface for dpkg (local .deb file installer)
type Client struct{}

// Name returns the package manager name
func (s *Client) Name() string {
	return "dpkg"
}

// Install installs local .deb files using dpkg -i
func (s *Client) Install(packages []string) error {
	args := append([]string{"-i"}, packages...)

	var cmd *exec.Cmd
	if osGetuid() != 0 {
		cmd = execCommand("sudo", append([]string{"dpkg"}, args...)...)
	} else {
		cmd = execCommand("dpkg", args...)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Remove is not supported by dpkg (use apt instead)
func (s *Client) Remove(_ []string) error {
	return errors.New("dpkg does not support package removal — use apt instead")
}

// Search lists installed packages matching the query using dpkg -l
// Output table format:
// ||/ Name           Version      Architecture Description
// +++-==============-============-============-=================================
// ii  package        version      arch         description
func (s *Client) Search(query string) ([]types.SearchResult, error) {
	cmd := execCommand("dpkg", "-l", "*"+query+"*")
	out, err := cmd.Output()
	if err != nil {
		// dpkg -l returns exit code 1 when no matches; treat as empty
		return nil, nil //nolint:nilerr // intentional: non-zero exit means no results
	}
	return parseDpkgSearch(string(out)), nil
}

// Update is not supported by dpkg (it is a local installer, not a repo manager)
func (s *Client) Update() error {
	return types.ErrNotSupported
}

// Upgrade is not supported by dpkg (it is a local installer, not a repo manager)
func (s *Client) Upgrade(packages []string) error {
	return types.ErrNotSupported
}

// List returns all installed packages using dpkg -l
func (s *Client) List() ([]types.PackageInfo, error) {
	cmd := execCommand("dpkg", "-l")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseDpkgList(string(out)), nil
}

// IsInstalled checks whether a package is installed using dpkg -s
func (s *Client) IsInstalled(pkg string) (bool, error) {
	cmd := execCommand("dpkg", "-s", pkg)
	err := cmd.Run()
	if err != nil {
		return false, nil
	}
	return true, nil
}

func parseDpkgList(output string) []types.PackageInfo {
	var results []types.PackageInfo
	scanner := bufio.NewScanner(strings.NewReader(output))
	headerPassed := false
	for scanner.Scan() {
		line := scanner.Text()
		// Skip header lines: both classic ("+++", "||/") and Unicode box-drawing ("├┼┼", "┌──")
		if strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "||/") ||
			strings.HasPrefix(line, "├") || strings.HasPrefix(line, "┌") ||
			strings.HasPrefix(line, "│") {
			headerPassed = true
			continue
		}
		if !headerPassed || strings.TrimSpace(line) == "" {
			continue
		}
		// Format: "ii  name  version  arch  description..."
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		// Only include installed packages (status "ii")
		if !strings.HasPrefix(fields[0], "ii") {
			continue
		}
		results = append(results, types.PackageInfo{
			Name:    fields[1],
			Version: fields[2],
		})
	}
	return results
}

func parseDpkgSearch(output string) []types.SearchResult {
	var results []types.SearchResult
	scanner := bufio.NewScanner(strings.NewReader(output))
	headerPassed := false
	for scanner.Scan() {
		line := scanner.Text()
		// Skip header lines: both classic ("+++", "||/") and Unicode box-drawing ("├┼┼", "┌──")
		if strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "||/") ||
			strings.HasPrefix(line, "├") || strings.HasPrefix(line, "┌") ||
			strings.HasPrefix(line, "│") {
			headerPassed = true
			continue
		}
		if !headerPassed || strings.TrimSpace(line) == "" {
			continue
		}
		// Format: "ii  name  version  arch  description..."
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		// Only show installed packages (status "ii")
		if !strings.HasPrefix(fields[0], "ii") {
			continue
		}
		results = append(results, types.SearchResult{
			Name:        fields[1],
			Version:     fields[2],
			Description: strings.Join(fields[4:], " "),
		})
	}
	return results
}
