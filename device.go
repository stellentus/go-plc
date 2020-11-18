package plc

import (
	"fmt"
)

// Device manages a connection to actual PLC hardware.
type Device struct {
	rawDevice
}

// NewDevice creates a new Device.
// The conConf string provides IP and other connection configuration (see libplctag for options).
// It is not thread safe. In a multi-threaded context, callers should ensure the appropriate
// portion of the tag tree is locked.
func NewDevice(conConf string, timeout int) (Device, error) {
	raw, err := newLibplctagDevice(conConf, timeout)
	if err != nil {
		return Device{}, err
	}
	return Device{rawDevice: &raw}, nil
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
	return dev.rawDevice.ReadTag(name, value)
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
