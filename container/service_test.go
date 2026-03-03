package container

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDetectorService(t *testing.T) {
	d := NewDetectorService()
	assert.NotNil(t, d)
}
