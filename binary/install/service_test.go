package install

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewInstallService_ReturnsNonNil(t *testing.T) {
	svc := NewInstallService()
	assert.IsType(t, &Service{}, svc)
}
