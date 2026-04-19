package install

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew_ReturnsNonNil(t *testing.T) {
	svc := New()
	assert.IsType(t, &Service{}, svc)
}
