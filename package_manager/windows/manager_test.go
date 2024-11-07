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
