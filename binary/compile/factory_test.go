package compile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBuildSystemByName(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"autotools", false},
		{"cmake", false},
		{"meson", false},
		{"makefile", false},
		{"cargo", false},
		{"golang", false},
		{"unknown", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs, err := NewBuildSystemByName(tt.name)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, bs)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, bs)
				assert.Equal(t, tt.name, bs.Name())
			}
		})
	}
}
