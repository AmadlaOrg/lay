package docker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewService(t *testing.T) {
	s := NewService()
	assert.NotNil(t, s)
}
