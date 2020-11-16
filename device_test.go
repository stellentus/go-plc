package plc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDevice(t *testing.T) {
	_, err := NewDevice("", 0)
	assert.NoError(t, err)
}
