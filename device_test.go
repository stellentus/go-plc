package plc

import (
	"fmt"
	"reflect"
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

func TestReadTagRequiresPointer(t *testing.T) {
	fake := RawDeviceFake{DeviceFake{}}
	dev := newTestDevice(&fake)

	var notAPointer int
	err := dev.ReadTag(testTagName, notAPointer)
	assert.Error(t, err)
}

func TestReadString(t *testing.T) {
	fake := RawDeviceFake{DeviceFake{
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
	fake := RawDeviceFake{DeviceFake{}}
	dev := newTestDevice(&fake)

	fake.DeviceFake[testTagName] = int(7)

	var result int
	err := dev.ReadTag(testTagName, &result)
	assert.NoError(t, err)

	assert.Equal(t, 7, result)
}

func TestWriteTag(t *testing.T) {
	fake := RawDeviceFake{DeviceFake{}}
	dev := newTestDevice(&fake)

	var value = 9
	err := dev.WriteTag(testTagName, value)
	assert.NoError(t, err)

	assert.Equal(t, 9, fake.DeviceFake[testTagName])
}

// RawDeviceFake adds lower APIs to a DeviceFake
type RawDeviceFake struct {
	DeviceFake
}

func (dev RawDeviceFake) Close() error {
	return nil
}

func (dev RawDeviceFake) GetList(listName, prefix string) ([]Tag, []string, error) {
	return nil, nil, nil
}

type DeviceFake map[string]interface{}

func (df DeviceFake) ReadTag(name string, value interface{}) error {
	v, ok := df[name]
	if !ok {
		return fmt.Errorf("")
	}

	in := reflect.ValueOf(v)
	out := reflect.Indirect(reflect.ValueOf(value))

	switch {
	case !out.CanSet():
		return fmt.Errorf("cannot set %s", out.Type().Name())
	case out.Kind() != in.Kind():
		return fmt.Errorf("cannot set %s to %s", out.Type().Name(), in.Type().Name())
	}

	out.Set(in)
	return nil
}

func (df DeviceFake) WriteTag(name string, value interface{}) error {
	df[name] = value
	return nil
}
