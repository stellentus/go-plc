package plc

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

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
	if !errors.Is(err, ErrBadRequest) {
		assert.Equal(t, err, ErrBadRequest, "Error should be of correct type")
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
	fake := FakeRawDevice{FakeReadWriter{}}
	dev := newTestDevice(&fake)

	var notAPointer int
	err := dev.ReadTag(testTagName, notAPointer)
	require.Error(t, err)

	if !errors.Is(err, ErrBadRequest) {
		require.Failf(t, "Incorrect error type for non-pointer read", "Received error: %v", err)
	}

	expectedError := ErrNonPointerRead{TagName: testTagName, Kind: reflect.Int}
	assert.Equal(t, err, expectedError, "Error should be of correct type")
}

func TestReadString(t *testing.T) {
	fake := FakeRawDevice{FakeReadWriter{
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
	fake := FakeRawDevice{FakeReadWriter{}}
	dev := newTestDevice(&fake)

	fake.FakeReadWriter[testTagName] = int(7)

	var result int
	err := dev.ReadTag(testTagName, &result)
	assert.NoError(t, err)

	assert.Equal(t, 7, result)
}

func TestWriteTag(t *testing.T) {
	fake := FakeRawDevice{FakeReadWriter{}}
	dev := newTestDevice(&fake)

	var value = 9
	err := dev.WriteTag(testTagName, value)
	assert.NoError(t, err)

	assert.Equal(t, 9, fake.FakeReadWriter[testTagName])
}

var _ = ReadWriter(FakeRawDevice{}) // Compiler makes sure this type is a ReadWriter
var _ = rawDevice(FakeRawDevice{})  // Compiler makes sure this type is a rawDevice

// FakeRawDevice adds lower APIs to a FakeReadWriter
type FakeRawDevice struct {
	FakeReadWriter
}

func (dev FakeRawDevice) Close() error {
	return nil
}

func (dev FakeRawDevice) GetList(listName, prefix string) ([]Tag, []string, error) {
	return nil, nil, nil
}

type FakeReadWriter map[string]interface{}

func (df FakeReadWriter) ReadTag(name string, value interface{}) error {
	v, ok := df[name]
	if !ok {
		return fmt.Errorf("FakeReadWriter does not contain '%s'", name)
	}

	in := reflect.ValueOf(v)
	out := reflect.Indirect(reflect.ValueOf(value))

	switch {
	case !out.CanSet():
		return fmt.Errorf("FakeReadWriter for '%s', cannot set %s", name, out.Type().Name())
	case out.Kind() != in.Kind():
		return fmt.Errorf("FakeReadWriter for '%s', cannot set %s to %s (%v)", name, out.Type().Name(), in.Type().Name(), v)
	}

	out.Set(in)
	return nil
}

func (df FakeReadWriter) WriteTag(name string, value interface{}) error {
	df[name] = value
	return nil
}

type fakeEvent struct {
	name string
	done chan interface{}
}

type BlockingFake struct {
	ch chan fakeEvent
	DeviceFake
}

func newBlockingFake() BlockingFake { return BlockingFake{make(chan fakeEvent), DeviceFake{}} }

func (bf BlockingFake) block(name string) {
	done := make(chan interface{})
	bf.ch <- fakeEvent{name, done}
	<-done
}

func (bf BlockingFake) ReadTag(name string, value interface{}) error {
	bf.block(name)
	return bf.DeviceFake.ReadTag(name, value)
}

func (bf BlockingFake) WriteTag(name string, value interface{}) error {
	bf.block(name)
	return bf.DeviceFake.WriteTag(name, value)
}
