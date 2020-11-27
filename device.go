package plc

import (
	"errors"
	"fmt"
	"reflect"
	"time"
)

// Device manages a connection to actual PLC hardware.
type Device struct {
	rawDevice
	conConf     string
	timeout     time.Duration
	isConnected bool
}

var _ = ReadWriter(&Device{}) // Compiler makes sure this type is a ReadWriter

// NewDevice creates a new Device at the provided address with options.
// It is not thread safe. In a multi-threaded context, callers should ensure the appropriate
// portion of the tag tree is locked.
func NewDevice(addr string, opts ...DeviceOptionFunc) (*Device, error) {
	if addr == "" {
		return nil, errors.New("Device cannot be initialized without an address")
	}

	dev := &Device{
		conConf: "gateway=" + addr,
	}

	for _, opt := range opts {
		err := opt(dev)
		if err != nil {
			return nil, err
		}
	}

	raw, err := newLibplctagDevice(dev.conConf, dev.timeout)
	if err != nil {
		return nil, err
	}

	dev.isConnected = true
	dev.rawDevice = &raw
	return dev, nil
}

type DeviceOptionFunc func(*Device) error

// Timeout sets the PLC connection timeout. Default is 5s.
func Timeout(to time.Duration) DeviceOptionFunc {
	return func(dev *Device) error {
		if dev.isConnected {
			return errors.New("Device timeout cannot be set after initialization")
		}
		dev.timeout = to
		return nil
	}
}

// LibplctagOption adds a libplctag option to the connection string (see libplctag for options).
// Here are some important ones:
// 	- protocol (default: "ab_eip")
// 	- path (default: "1,0")
// 	- cpu (default: "controllogix")
func LibplctagOption(name, val string) DeviceOptionFunc {
	return func(dev *Device) error {
		if dev.isConnected {
			return errors.New("Libplctag options cannot be set after initialization")
		}
		if name == "" {
			return errors.New("Libplctag option name was not set")
		}
		dev.conConf += "&" + name + "=" + val
		return nil
	}
}

// Close should be called on the Device to clean up its resources.
func (dev *Device) Close() error {
	return dev.rawDevice.Close()
}

// TagWithIndex provides the fully qualified tag for the given index of an array.
func TagWithIndex(name string, index int) string {
	// Array tags can be read by adding the index to the string, e.g. "EXAMPLE[0]"
	// Perhaps this should have error checking on index<0.
	return fmt.Sprintf("%s[%d]", name, index)
}

// ReadTag reads the requested tag into the provided value.
// It is not thread safe. In a multi-threaded context, callers should ensure the appropriate
// portion of the tag tree is locked.
func (dev *Device) ReadTag(name string, value interface{}) error {
	v := reflect.ValueOf(value)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("ReadTag expects a pointer type but got %v", v.Kind())
	}

	switch v.Elem().Kind() {
	case reflect.String:
		bytes := make([]byte, stringMaxLength)
		for str_index := 0; str_index < stringMaxLength; str_index++ {
			var val byte
			err := dev.rawDevice.ReadTag(TagWithIndex(name, str_index), &val)
			if err != nil {
				return err
			}
			if val == 0 {
				// We found a null, which is the end of the string
				bytes = bytes[:str_index] // we don't want the nulls at the end
				break
			}
			bytes[str_index] = val
		}
		result := string(bytes)
		v.Elem().Set(reflect.ValueOf(result))
		return nil
	default:
		return dev.rawDevice.ReadTag(name, value)
	}
}

// WriteTag writes the provided tag and value.
// It is not thread safe. In a multi-threaded context, callers should ensure the appropriate
// portion of the tag tree is locked.
func (dev *Device) WriteTag(name string, value interface{}) error {
	return dev.rawDevice.WriteTag(name, value)
}

// GetAllTags gets a list of all tags available on the Device.
func (dev *Device) GetAllTags() ([]Tag, error) {
	tags, programs, err := dev.rawDevice.GetList("", "")
	if err != nil {
		return nil, err
	}

	for _, progName := range programs {
		progTags, _, err := dev.rawDevice.GetList(progName, "")
		if err != nil {
			return nil, err
		}
		tags = append(tags, progTags...)
	}

	return tags, nil
}

// GetAllPrograms gets a list of all programs on the Device.
func (dev *Device) GetAllPrograms() ([]string, error) {
	_, programs, err := dev.rawDevice.GetList("", "")
	if err != nil {
		return nil, err
	}

	return programs, nil
}
