package linux

import (
	"testing"
)

func TestValuesEnum(t *testing.T) {
	tests := []struct {
		name     string
		value    Values
		expected string
	}{
		{
			name:     "apt value",
			value:    aptVal,
			expected: "apt",
		},
		{
			name:     "dnf value",
			value:    dnfVal,
			expected: "dnf",
		},
		{
			name:     "dpkg value",
			value:    dpkgVal,
			expected: "dpkg",
		},
		{
			name:     "rpm value",
			value:    rpmVal,
			expected: "rpm",
		},
		{
			name:     "yum value",
			value:    yumVal,
			expected: "yum",
		},
		{
			name:     "pacman value",
			value:    pacmanVal,
			expected: "pacman",
		},
		{
			name:     "zypper value",
			value:    zypperVal,
			expected: "zypper",
		},
		{
			name:     "apk value",
			value:    apkVal,
			expected: "apk",
		},
		{
			name:     "nix value",
			value:    nixVal,
			expected: "nix",
		},
		{
			name:     "snap value",
			value:    snapVal,
			expected: "snap",
		},
		{
			name:     "flatpak value",
			value:    flatpakVal,
			expected: "flatpak",
		},
		{
			name:     "Invalid value",
			value:    Values(20),
			expected: "Unknown(20)",
		},
		{
			name:     "Negative invalid value",
			value:    Values(-1),
			expected: "Unknown(-1)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.value.String()

			if got != tt.expected {
				t.Errorf("Expected %v, but got %v", tt.expected, got)
			}
		})
	}
}

func TestExists(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"apt exists", "apt", true},
		{"dnf exists", "dnf", true},
		{"pacman exists", "pacman", true},
		{"zypper exists", "zypper", true},
		{"apk exists", "apk", true},
		{"nix exists", "nix", true},
		{"snap exists", "snap", true},
		{"flatpak exists", "flatpak", true},
		{"dpkg exists", "dpkg", true},
		{"rpm exists", "rpm", true},
		{"unknown doesn't exist", "brew", false},
		{"empty doesn't exist", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Exists(tt.input)
			if got != tt.expected {
				t.Errorf("Exists(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}
