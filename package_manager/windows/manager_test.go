package windows

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
			name:     "choco value",
			value:    choco,
			expected: "choco",
		},
		{
			name:     "scoop value",
			value:    scoop,
			expected: "scoop",
		},
		{
			name:     "winget value",
			value:    winget,
			expected: "winget",
		},
		{
			name:     "Invalid value",
			value:    Values(8),
			expected: "Unknown(8)",
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
		{"choco exists", "choco", true},
		{"scoop exists", "scoop", true},
		{"winget exists", "winget", true},
		{"unknown doesn't exist", "apt", false},
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
