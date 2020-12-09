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
	timeout time.Duration
	conf    map[string]string
}

var _ = ReadWriter(&Device{}) // Compiler makes sure this type is a ReadWriter

var (
	ErrNoAddress = errors.New("Device cannot be initialized without an address")
)

// NewDevice creates a new Device at the provided address with options.
// It is not thread safe. In a multi-threaded context, callers should ensure the appropriate
// portion of the tag tree is locked.
func NewDevice(addr string, opts ...deviceOption) (*Device, error) {
	if addr == "" {
		return nil, ErrNoAddress
	}

	// Initialize with default connection options
	dev := &Device{
		conf: map[string]string{
			"protocol": "ab_eip",
			"path":     "1,0",
			"cpu":      "controllogix",
		},
		timeout: 5 * time.Second,
	}

	for _, opt := range opts {
		opt(dev)
	}

	conConf := "gateway=" + addr
	for name, val := range dev.conf {
		conConf += "&" + name + "=" + val
	}

	dev.rawDevice = newLibplctagDevice(conConf, dev.timeout)
	return dev, nil
}

type deviceOption func(*Device)

// Timeout sets the PLC connection timeout. Default is 5s.
func Timeout(to time.Duration) deviceOption {
	return func(dev *Device) {
		dev.timeout = to
	}
}

// LibplctagOption adds a libplctag option to the connection string (see libplctag for options).
// Here are some important ones:
// 	- protocol (default: "ab_eip")
// 	- path (default: "1,0")
// 	- cpu (default: "controllogix")
func LibplctagOption(name, val string) deviceOption {
	return func(dev *Device) {
		dev.conf[name] = val
	}
}

// Close should be called on the Device to clean up its resources.
func (dev *Device) Close() error {
	err := dev.rawDevice.Close()
	if err != nil {
		return fmt.Errorf("device close: %w", err)
	}
	return nil
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
		return newErrNonPointerRead(name, v.Kind())
	}

	switch v.Elem().Kind() {
	case reflect.String:
		bytes := make([]byte, stringMaxLength)
		for str_index := 0; str_index < stringMaxLength; str_index++ {
			var val byte
			tagWithIndex := TagWithIndex(name, str_index)
			err := dev.rawDevice.ReadTag(tagWithIndex, &val)
			if err != nil {
				return fmt.Errorf("ReadTag '%s' as string: %w", tagWithIndex, err)
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
	default:
		err := dev.rawDevice.ReadTag(name, value)
		if err != nil {
			return fmt.Errorf("ReadTag '%s': %w", name, err)
		}
	}

	return nil
}

// WriteTag writes the provided tag and value.
// It is not thread safe. In a multi-threaded context, callers should ensure the appropriate
// portion of the tag tree is locked.
func (dev *Device) WriteTag(name string, value interface{}) error {
	err := dev.rawDevice.WriteTag(name, value)
	if err != nil {
		return fmt.Errorf("WriteTag '%s': %w", name, err)
	}
	return nil
}

// GetAllTags gets a list of all tags available on the Device.
func (dev *Device) GetAllTags() ([]Tag, error) {
	tags, programs, err := dev.rawDevice.GetList("", "")
	if err != nil {
		return nil, fmt.Errorf("GetAllTags: %w", err)
	}

	for _, progName := range programs {
		progTags, _, err := dev.rawDevice.GetList(progName, "")
		if err != nil {
			return nil, fmt.Errorf("GetAllTags for program '%s': %w", progName, err)
		}
		tags = append(tags, progTags...)
	}

	return tags, nil
}

// GetAllPrograms gets a list of all programs on the Device.
func (dev *Device) GetAllPrograms() ([]string, error) {
	_, programs, err := dev.rawDevice.GetList("", "")
	if err != nil {
		return nil, fmt.Errorf("GetAllPrograms: %w", err)
	}

	return programs, nil
}

type ErrNonPointerRead struct {
	TagName string
	reflect.Kind
}

func newErrNonPointerRead(tn string, k reflect.Kind) ErrNonPointerRead {
	return ErrNonPointerRead{TagName: tn, Kind: k}
}

func (err ErrNonPointerRead) Error() string {
	return fmt.Sprintf("ReadTag expects a pointer type but got %v for tag '%s'", err.Kind, err.TagName)
}
