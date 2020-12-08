package plc

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewDevice(t *testing.T) {
	_, err := NewDevice("test")
	assert.NoError(t, err)
}

func TestNewDeviceRequiresGateway(t *testing.T) {
	_, err := NewDevice("")
	assert.Error(t, err)
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
	assert.Error(t, err)
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
