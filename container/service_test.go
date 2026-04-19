package container

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	d := New()
	assert.NotNil(t, d)
}
