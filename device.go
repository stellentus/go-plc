package plc

/*
#include <stdlib.h>
#include "./libplctag.h"
*/
import "C"
import (
	"fmt"
	"strings"
	"unsafe"
)

// Device manages a connection to actual PLC hardware.
type Device struct {
	conConf string
	ids     map[string]C.int32_t
	timeout C.int
}

// NewDevice creates a new Device.
// The conConf string provides IP and other connection configuration (see libplctag for options).
func NewDevice(conConf string, timeout int) (Device, error) {
	dev := Device{
		conConf: conConf,
		ids:     make(map[string]C.int32_t),
		timeout: C.int(timeout),
	}

	return dev, nil
}

const noOffset = C.int(0)

func (dev *Device) getID(tagName string) (C.int32_t, error) {
	id, ok := dev.ids[tagName]
	if ok {
		return id, nil
	}

	cattrib_str := C.CString(dev.conConf + "&name=" + tagName) // can also specify elem_size=1&elem_count=1
	defer C.free(unsafe.Pointer(cattrib_str))

	id = C.plc_tag_create(cattrib_str, dev.timeout)
	if id < 0 {
		return id, newError(id)
	}
	dev.ids[tagName] = id
	return id, nil
}

// Close should be called on the Device to clean up its resources.
func (dev *Device) Close() error {
	for _, id := range dev.ids {
		err := newError(C.plc_tag_destroy(id))
		if err != nil {
			return err
		}
	}
	return nil
}

// StatusForTag returns the error status of the requested tag.
func (dev *Device) StatusForTag(name string) error {
	id, err := dev.getID(name)
	if err != nil {
		return err
	}
	return newError(C.plc_tag_status(id))
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
	id, err := dev.getID(name)
	if err != nil {
		return err
	}

	if err := newError(C.plc_tag_read(id, dev.timeout)); err != nil {
		return err
	}

	switch val := value.(type) {
	case *bool:
		result := C.plc_tag_get_uint8(id, noOffset)
		*val = uint8(result) > 0
	case *uint8:
		result := C.plc_tag_get_uint8(id, noOffset)
		*val = uint8(result)
	case *uint16:
		result := C.plc_tag_get_uint16(id, noOffset)
		*val = uint16(result)
	case *uint32:
		result := C.plc_tag_get_uint32(id, noOffset)
		*val = uint32(result)
	case *uint64:
		result := C.plc_tag_get_uint64(id, noOffset)
		*val = uint64(result)
	case *int8:
		result := C.plc_tag_get_int8(id, noOffset)
		*val = int8(result)
	case *int16:
		result := C.plc_tag_get_int16(id, noOffset)
		*val = int16(result)
	case *int32:
		result := C.plc_tag_get_int32(id, noOffset)
		*val = int32(result)
	case *int64:
		result := C.plc_tag_get_int64(id, noOffset)
		*val = int64(result)
	case *float32:
		result := C.plc_tag_get_float32(id, noOffset)
		*val = float32(result)
	case *float64:
		result := C.plc_tag_get_float64(id, noOffset)
		*val = float64(result)
	default:
		return fmt.Errorf("Type %T is unknown and can't be read (%v)", val, val)
	}

	return nil
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
	id, err := dev.getID(name)
	if err != nil {
		return err
	}

	switch val := value.(type) {
	case bool:
		b := C.uint8_t(0)
		if val {
			b = C.uint8_t(255)
		}
		err = newError(C.plc_tag_set_uint8(id, noOffset, b))
	case uint8:
		err = newError(C.plc_tag_set_uint8(id, noOffset, C.uint8_t(val)))
	case uint16:
		err = newError(C.plc_tag_set_uint16(id, noOffset, C.uint16_t(val)))
	case uint32:
		err = newError(C.plc_tag_set_uint32(id, noOffset, C.uint32_t(val)))
	case uint64:
		err = newError(C.plc_tag_set_uint64(id, noOffset, C.uint64_t(val)))
	case int8:
		err = newError(C.plc_tag_set_int8(id, noOffset, C.int8_t(val)))
	case int16:
		err = newError(C.plc_tag_set_int16(id, noOffset, C.int16_t(val)))
	case int32:
		err = newError(C.plc_tag_set_int32(id, noOffset, C.int32_t(val)))
	case int64:
		err = newError(C.plc_tag_set_int64(id, noOffset, C.int64_t(val)))
	case float32:
		err = newError(C.plc_tag_set_float32(id, noOffset, C.float(val)))
	case float64:
		err = newError(C.plc_tag_set_float64(id, noOffset, C.double(val)))
	default:
		err = fmt.Errorf("Type %T is unknown and can't be written (%v)", val, val)
	}
	if err != nil {
		return err
	}

	// Read. If non-zero, value is true. Otherwise, it's false.
	if err := newError(C.plc_tag_write(id, dev.timeout)); err != nil {
		return err
	}

	return nil
}

// GetAllTags gets a list of all tags available on the Device.
func (dev *Device) GetAllTags() ([]Tag, error) {
	tags, programs, err := dev.getList("", "")
	if err != nil {
		return nil, err
	}

	for _, progName := range programs {
		progTags, _, err := dev.getList(progName, "")
		if err != nil {
			return nil, err
		}
		tags = append(tags, progTags...)
	}

	return tags, nil
}

// GetAllPrograms gets a list of all programs on the Device.
func (dev *Device) GetAllPrograms() ([]string, error) {
	_, programs, err := dev.getList("", "")
	if err != nil {
		return nil, err
	}

	return programs, nil
}

func (dev *Device) getList(listName, prefix string) ([]Tag, []string, error) {
	if listName == "" {
		listName += "@tags"
	} else {
		listName += ".@tags"
	}

	id, err := dev.getID(listName)
	if err != nil {
		return nil, nil, err
	}

	if err := newError(C.plc_tag_read(id, dev.timeout)); err != nil {
		return nil, nil, err
	}

	tags := []Tag{}
	programNames := []string{}

	offset := C.int(0)
	for {
		tag := Tag{}
		offset += 4

		tag.tagType = C.plc_tag_get_uint16(id, offset)
		offset += 2

		tag.elementSize = C.plc_tag_get_uint16(id, offset)
		offset += 2

		tag.addDimension(int(C.plc_tag_get_uint32(id, offset)))
		offset += 4
		tag.addDimension(int(C.plc_tag_get_uint32(id, offset)))
		offset += 4
		tag.addDimension(int(C.plc_tag_get_uint32(id, offset)))
		offset += 4

		nameLength := int(C.plc_tag_get_uint16(id, offset))
		offset += 2

		tagBytes := make([]byte, nameLength)
		for i := 0; i < nameLength; i++ {
			tagBytes[i] = byte(C.plc_tag_get_int8(id, offset))
			offset++
		}

		if prefix != "" {
			tag.name = prefix + "." + string(tagBytes)
		} else {
			tag.name = string(tagBytes)
		}

		if strings.HasPrefix(tag.name, "Program:") {
			programNames = append(programNames, tag.name)
		} else if (tag.tagType & SystemTagBit) == SystemTagBit {
			// Do nothing for system tags
		} else {
			numDimensions := int((tag.tagType & TagDimensionMask) >> 13)
			if numDimensions != len(tag.dimensions) {
				return nil, nil, fmt.Errorf("Tag '%s' claims to have %d dimensions but has %d", tag.name, numDimensions, len(tag.dimensions))
			}

			tags = append(tags, tag)
		}

		if offset >= C.plc_tag_get_size(id) {
			break
		}
	}

	return tags, programNames, nil
}
