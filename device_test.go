package plc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDevice(t *testing.T) {
	_, err := NewDevice("", 0)
	assert.NoError(t, err)
}

func newTestDevice(rd rawDevice) Device {
	return Device{rawDevice: rd}
}

const testTagName = "TEST_TAG"

func TestReadTagAtIndex(t *testing.T) {
	spy := spyRawDevice{}
	dev := newTestDevice(&spy)

	var unused int
	err := dev.ReadTagAtIndex(testTagName, 0, &unused)
	assert.NoError(t, err)

	assert.Equal(t, testTagName+"[0]", spy.lastTag)
}

func TestWriteTagAtIndex(t *testing.T) {
	spy := spyRawDevice{}
	dev := newTestDevice(&spy)

	var unused int
	err := dev.WriteTagAtIndex(testTagName, 0, &unused)
	assert.NoError(t, err)

	assert.Equal(t, testTagName+"[0]", spy.lastTag)
}

// spyRawDevice just records the last tag that was sent through the interface
type spyRawDevice struct {
	lastTag string
}

func (dev *spyRawDevice) ReadTag(name string, value interface{}) error {
	dev.lastTag = name
	return nil
}

func (dev *spyRawDevice) WriteTag(name string, value interface{}) error {
	dev.lastTag = name
	return nil
}

func (dev *spyRawDevice) Close() error {
	return nil
}

func (dev *spyRawDevice) StatusForTag(name string) error {
	dev.lastTag = name
	return nil
}

func (dev *spyRawDevice) GetList(listName, prefix string) ([]Tag, []string, error) {
	return nil, nil, nil
}
