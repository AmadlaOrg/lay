package nix

import (
	"os"
	"os/exec"
	"strings"

	"github.com/AmadlaOrg/lay/package_manager/types"
)

// For mocking
var execCommand = exec.Command

// Client implements the package manager interface for nix
type Client struct{}

// Name returns the package manager name
func (s *Client) Name() string {
	return "nix"
}

// Install installs packages using nix profile install
func (s *Client) Install(packages []string) error {
	// nix profile install doesn't need sudo — it installs to user profile
	// Packages are prefixed with nixpkgs# if not already qualified
	var args []string
	args = append(args, "profile", "install")
	for _, pkg := range packages {
		if !strings.Contains(pkg, "#") {
			pkg = "nixpkgs#" + pkg
		}
		args = append(args, pkg)
	}

	cmd := execCommand("nix", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Remove removes packages using nix profile remove
func (s *Client) Remove(packages []string) error {
	args := append([]string{"-e"}, packages...)
	cmd := execCommand("nix-env", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Search searches for packages using nix search
// Output format:
// * nixpkgs#package (version)
//
//	description
func (s *Client) Search(query string) ([]types.SearchResult, error) {
	cmd := execCommand("nix", "search", "nixpkgs", query, "--no-update-lock-file")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseNixSearch(string(out)), nil
}

func parseNixSearch(output string) []types.SearchResult {
	var results []types.SearchResult
	lines := strings.Split(output, "\n")
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if !strings.HasPrefix(line, "* ") {
			continue
		}
		// Format: "* nixpkgs#name (version)"
		header := strings.TrimPrefix(line, "* ")
		var name, version string

		if parenIdx := strings.Index(header, " ("); parenIdx >= 0 {
			name = header[:parenIdx]
			version = strings.Trim(header[parenIdx+2:], ")")
		} else {
			name = header
		}

		// Strip nixpkgs# prefix
		if idx := strings.Index(name, "#"); idx >= 0 {
			name = name[idx+1:]
		}

		var desc string
		if i+1 < len(lines) {
			desc = strings.TrimSpace(lines[i+1])
			i++
		}

		results = append(results, types.SearchResult{
			Name:        name,
			Version:     version,
			Description: desc,
		})
	}
	return results
}

// Update runs nix flake update to refresh the flake lock file
func (s *Client) Update() error {
	cmd := execCommand("nix", "flake", "update")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Upgrade upgrades packages. If packages is empty, upgrades all installed packages.
// If packages are specified, upgrades them one at a time.
func (s *Client) Upgrade(packages []string) error {
	if len(packages) == 0 {
		cmd := execCommand("nix", "profile", "upgrade", ".*")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		return cmd.Run()
	}

	for _, pkg := range packages {
		cmd := execCommand("nix", "profile", "upgrade", pkg)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

// List returns all installed packages from the user profile
// Output format: "Index FlakeRef ResolvedRef StorePath"
func (s *Client) List() ([]types.PackageInfo, error) {
	cmd := execCommand("nix", "profile", "list")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseNixList(string(out)), nil
}

func parseNixList(output string) []types.PackageInfo {
	var results []types.PackageInfo
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Format: "Index FlakeRef ResolvedRef StorePath"
		// Example: "0 flake:nixpkgs#curl github:NixOS/nixpkgs/abc123#curl /nix/store/xxx-curl-8.5.0"
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		name := fields[1]
		// Strip "flake:nixpkgs#" or similar prefix to get the package name
		if idx := strings.Index(name, "#"); idx >= 0 {
			name = name[idx+1:]
		}
		results = append(results, types.PackageInfo{
			Name:    name,
			Version: "",
		})
	}
	return results
}

// IsInstalled checks whether a package is installed by searching nix profile list output
func (s *Client) IsInstalled(pkg string) (bool, error) {
	cmd := execCommand("nix", "profile", "list")
	out, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return strings.Contains(string(out), pkg), nil
}
