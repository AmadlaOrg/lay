package compile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	svc := New()
	assert.NotNil(t, svc)
}
