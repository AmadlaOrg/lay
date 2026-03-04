package package_manager

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewManagerByName(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectName  string
		expectErr   bool
		errContains string
	}{
		{
			name:       "apt returns apt manager",
			input:      "apt",
			expectName: "apt",
		},
		{
			name:       "dnf returns dnf manager",
			input:      "dnf",
			expectName: "dnf",
		},
		{
			name:       "yum returns yum manager",
			input:      "yum",
			expectName: "yum",
		},
		{
			name:       "pacman returns pacman manager",
			input:      "pacman",
			expectName: "pacman",
		},
		{
			name:       "zypper returns zypper manager",
			input:      "zypper",
			expectName: "zypper",
		},
		{
			name:       "apk returns apk manager",
			input:      "apk",
			expectName: "apk",
		},
		{
			name:       "nix returns nix manager",
			input:      "nix",
			expectName: "nix",
		},
		{
			name:       "snap returns snap manager",
			input:      "snap",
			expectName: "snap",
		},
		{
			name:       "flatpak returns flatpak manager",
			input:      "flatpak",
			expectName: "flatpak",
		},
		{
			name:       "dpkg returns dpkg manager",
			input:      "dpkg",
			expectName: "dpkg",
		},
		{
			name:       "rpm returns rpm manager",
			input:      "rpm",
			expectName: "rpm",
		},
		{
			name:       "choco returns choco manager",
			input:      "choco",
			expectName: "choco",
		},
		{
			name:       "scoop returns scoop manager",
			input:      "scoop",
			expectName: "scoop",
		},
		{
			name:       "winget returns winget manager",
			input:      "winget",
			expectName: "winget",
		},
		{
			name:       "brew returns brew manager",
			input:      "brew",
			expectName: "brew",
		},
		{
			name:        "unknown name returns error",
			input:       "yast",
			expectErr:   true,
			errContains: "unsupported package manager: yast",
		},
		{
			name:        "empty name returns error",
			input:       "",
			expectErr:   true,
			errContains: "unsupported package manager",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm, err := NewManagerByName(tt.input)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, pm)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, pm)
				assert.Equal(t, tt.expectName, pm.Name())
			}
		})
	}
}
