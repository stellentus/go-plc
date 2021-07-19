package physical

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/stellentus/go-plc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDevice(t *testing.T) {
	_, err := NewDevice("test")
	assert.NoError(t, err)
}

func TestNewDeviceRequiresAddress(t *testing.T) {
	_, err := NewDevice("")
	require.Error(t, err)

	// Ideally we'd call assert.IsError, but it's not available yet.
	// All we really care about is `errors.Is`, not strict equality.
	// So use errors.Is for the soft check, but use assert.Equal for pretty printing.
	if !errors.Is(err, plc.ErrBadRequest) {
		assert.Equal(t, err, plc.ErrBadRequest, "Error should be of correct type")
	}
}

func TestNewDeviceParsesTimeout(t *testing.T) {
	dev, err := NewDevice("test", Timeout(time.Millisecond))
	assert.NoError(t, err)
	assert.Equal(t, time.Millisecond, dev.timeout)
}

func newTestDevice(rd rawDevice) *Device {
	return &Device{rawDevice: rd}
}

const testTagName = "TEST_TAG"

func TestReadTagRequiresPointer(t *testing.T) {
	fake := FakeRawDevice{plc.FakeReadWriter{}}
	dev := newTestDevice(&fake)

	var notAPointer int
	err := dev.ReadTag(testTagName, notAPointer)
	require.Error(t, err)

	if !errors.Is(err, plc.ErrBadRequest) {
		require.Failf(t, "Incorrect error type for non-pointer read", "Received error: %v", err)
	}

	expectedError := plc.ErrNonPointerRead{TagName: testTagName, Kind: reflect.Int}
	assert.Equal(t, err, expectedError, "Error should be of correct type")
}

func TestReadString(t *testing.T) {
	fake := FakeRawDevice{plc.FakeReadWriter{
		"STR[0]": uint8('h'),
		"STR[1]": uint8('i'),
		"STR[2]": uint8(0),
	}}
	dev := newTestDevice(&fake)

	var str string
	err := dev.ReadTag("STR", &str)
	assert.NoError(t, err)
	assert.Equal(t, "hi", str, "String should be loaded from array elements ending in null")
}

func TestReadTag(t *testing.T) {
	fake := FakeRawDevice{plc.FakeReadWriter{}}
	dev := newTestDevice(&fake)

	fake.FakeReadWriter[testTagName] = int(7)

	var result int
	err := dev.ReadTag(testTagName, &result)
	assert.NoError(t, err)

	assert.Equal(t, 7, result)
}

func TestWriteTag(t *testing.T) {
	fake := FakeRawDevice{plc.FakeReadWriter{}}
	dev := newTestDevice(&fake)

	var value = 9
	err := dev.WriteTag(testTagName, value)
	assert.NoError(t, err)

	assert.Equal(t, 9, fake.FakeReadWriter[testTagName])
}

var _ = plc.ReadWriter(FakeRawDevice{}) // Compiler makes sure this type is a ReadWriter
var _ = rawDevice(FakeRawDevice{})      // Compiler makes sure this type is a rawDevice

// FakeRawDevice adds lower APIs to a plc.FakeReadWriter
type FakeRawDevice struct {
	plc.FakeReadWriter
}

func (dev FakeRawDevice) Close() error {
	return nil
}

func (dev FakeRawDevice) GetList(listName, prefix string) ([]plc.Tag, []string, error) {
	return nil, nil, nil
}
