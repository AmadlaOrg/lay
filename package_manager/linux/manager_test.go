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
			value:    apt,
			expected: "apt",
		},
		{
			name:     "dnf value",
			value:    dnf,
			expected: "dnf",
		},
		{
			name:     "dpkg value",
			value:    dpkg,
			expected: "dpkg",
		},
		{
			name:     "rpm value",
			value:    rpm,
			expected: "rpm",
		},
		{
			name:     "yum value",
			value:    yum,
			expected: "yum",
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
