package package_manager

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDetectorService(t *testing.T) {
	svc := NewDetectorService()
	assert.NotNil(t, svc)
}
