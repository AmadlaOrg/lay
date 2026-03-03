package compile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectorService_Detect(t *testing.T) {
	tests := []struct {
		name     string
		markers  []string
		override string
		expected string
		wantErr  bool
	}{
		{
			name:     "override takes priority",
			override: "cargo",
			expected: "cargo",
		},
		{
			name:     "detects configure (autotools)",
			markers:  []string{"configure"},
			expected: "autotools",
		},
		{
			name:     "detects configure.ac (autotools)",
			markers:  []string{"configure.ac"},
			expected: "autotools",
		},
		{
			name:     "detects CMakeLists.txt",
			markers:  []string{"CMakeLists.txt"},
			expected: "cmake",
		},
		{
			name:     "detects meson.build",
			markers:  []string{"meson.build"},
			expected: "meson",
		},
		{
			name:     "detects Cargo.toml",
			markers:  []string{"Cargo.toml"},
			expected: "cargo",
		},
		{
			name:     "detects go.mod",
			markers:  []string{"go.mod"},
			expected: "golang",
		},
		{
			name:     "detects Makefile (lowest priority)",
			markers:  []string{"Makefile"},
			expected: "makefile",
		},
		{
			name:     "autotools wins over Makefile",
			markers:  []string{"configure", "Makefile"},
			expected: "autotools",
		},
		{
			name:    "no markers found",
			markers: []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			for _, m := range tt.markers {
				_ = os.WriteFile(filepath.Join(dir, m), []byte(""), 0644)
			}

			s := &DetectorService{}
			result, err := s.Detect(dir, tt.override)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
