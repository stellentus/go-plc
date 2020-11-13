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

// StatusForTag returns the error status of the requested tag.
func (dev *Device) StatusForTag(name string) error {
	return dev.rawDevice.StatusForTag(name)
}

func tagWithIndex(name string, index int) string {
	// Array tags can be read by adding the index to the string, e.g. "EXAMPLE[0]"
	// Perhaps this should have error checking on index<0.
	return fmt.Sprintf("%s[%d]", name, index)
}

// ReadTagAtIndex reads the requested array tag at the given index into the provided value.
// It's provided to be faster than ReadTag when only a single array element is needed.
func (dev *Device) ReadTagAtIndex(name string, index int, value interface{}) error {
	name = tagWithIndex(name, index)
	return dev.ReadTag(name, value)
}

// ReadTag reads the requested tag into the provided value.
func (dev *Device) ReadTag(name string, value interface{}) error {
	return dev.rawDevice.ReadTag(name, value)
}

// WriteTagAtIndex writes the requested array tag at the given index with the provided value.
// It's provided to be faster than WriteTag when only a single array element is needed. (Otherwise
// would be necessary to read into an entire slice, edit one element, and re-write the slice,
// which is not atomic.)
func (dev *Device) WriteTagAtIndex(name string, index int, value interface{}) error {
	name = tagWithIndex(name, index)
	return dev.WriteTag(name, value)
}

// WriteTag writes the provided tag and value.
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
