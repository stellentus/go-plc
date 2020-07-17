package plc

/*
#cgo LDFLAGS: ./libplctag.a
#include <stdio.h>
#include <stdlib.h>
#include "./libplctag.h"
*/
import "C"
import (
	"fmt"
	"unsafe"
)

type PLC struct {
	conConf string
	ids     map[string]C.int32_t
	timeout C.int
}

// New creates a new PLC.
func New(conConf string, timeout int) (PLC, error) {
	plc := PLC{
		conConf: conConf,
		ids:     make(map[string]C.int32_t),
		timeout: C.int(timeout),
	}

	return plc, nil
}

func (plc *PLC) getID(tagName string) (C.int32_t, error) {
	id, ok := plc.ids[tagName]
	if ok {
		return id, nil
	}

	cattrib_str := C.CString(plc.conConf + "&name=" + tagName) // can also specify elem_size=1&elem_count=1
	defer C.free(unsafe.Pointer(cattrib_str))

	id = C.plc_tag_create(cattrib_str, plc.timeout)
	if id < 0 {
		return id, newError(id)
	}
	return id, nil
}

// Close should be called on the PLC to clean up its resources.
func (plc *PLC) Close() error {
	for _, id := range plc.ids {
		err := newError(C.plc_tag_destroy(id))
		if err != nil {
			return err
		}
	}
	return nil
}

// StatusForTag returns the error status of the requested tag.
func (plc *PLC) StatusForTag(name string) error {
	id, err := plc.getID(name)
	if err != nil {
		return err
	}
	return newError(C.plc_tag_status(id))
}

// ReadTag reads the requested tag into the provided value.
func (plc *PLC) ReadTag(name string, value interface{}) error {
	id, err := plc.getID(name)
	if err != nil {
		return err
	}

	if err := newError(C.plc_tag_read(id, plc.timeout)); err != nil {
		return err
	}

	offset := C.int(0)

	switch val := value.(type) {
	case *bool:
		result := C.plc_tag_get_uint8(id, offset)
		*val = uint8(result) > 0
	case *uint8:
		result := C.plc_tag_get_uint8(id, offset)
		*val = uint8(result)
	case *uint16:
		result := C.plc_tag_get_uint16(id, offset)
		*val = uint16(result)
	case *uint32:
		result := C.plc_tag_get_uint32(id, offset)
		*val = uint32(result)
	case *uint64:
		result := C.plc_tag_get_uint64(id, offset)
		*val = uint64(result)
	case *int8:
		result := C.plc_tag_get_int8(id, offset)
		*val = int8(result)
	case *int16:
		result := C.plc_tag_get_int16(id, offset)
		*val = int16(result)
	case *int32:
		result := C.plc_tag_get_int32(id, offset)
		*val = int32(result)
	case *int64:
		result := C.plc_tag_get_int64(id, offset)
		*val = int64(result)
	case *float32:
		result := C.plc_tag_get_float32(id, offset)
		*val = float32(result)
	case *float64:
		result := C.plc_tag_get_float64(id, offset)
		*val = float64(result)
	default:
		fmt.Printf("Type %T is unknown and can't be read (%v)\n", val, val)
	}

	return nil
}

// WriteTag writes the provided tag and value.
func (plc *PLC) WriteTag(name string, value interface{}) error {
	id, err := plc.getID(name)
	if err != nil {
		return err
	}

	offset := C.int(0)

	switch val := value.(type) {
	case bool:
		b := C.uint8_t(0)
		if val {
			b = C.uint8_t(255)
		}
		err = newError(C.plc_tag_set_uint8(id, offset, b))
	case uint8:
		err = newError(C.plc_tag_set_uint8(id, offset, C.uint8_t(val)))
	case uint16:
		err = newError(C.plc_tag_set_uint16(id, offset, C.uint16_t(val)))
	case uint32:
		err = newError(C.plc_tag_set_uint32(id, offset, C.uint32_t(val)))
	case uint64:
		err = newError(C.plc_tag_set_uint64(id, offset, C.uint64_t(val)))
	case int8:
		err = newError(C.plc_tag_set_int8(id, offset, C.int8_t(val)))
	case int16:
		err = newError(C.plc_tag_set_int16(id, offset, C.int16_t(val)))
	case int32:
		err = newError(C.plc_tag_set_int32(id, offset, C.int32_t(val)))
	case int64:
		err = newError(C.plc_tag_set_int64(id, offset, C.int64_t(val)))
	case float32:
		err = newError(C.plc_tag_set_float32(id, offset, C.float(val)))
	case float64:
		err = newError(C.plc_tag_set_float64(id, offset, C.double(val)))
	default:
		fmt.Printf("Type %T is unknown and can't be written (%v)\n", val, val)
	}
	if err != nil {
		return err
	}

	// Read. If non-zero, value is true. Otherwise, it's false.
	if err := newError(C.plc_tag_write(id, plc.timeout)); err != nil {
		return err
	}

	return nil
}

func CheckRequiredVersion(major, minor, patch int) error {
	return newError(C.plc_tag_check_lib_version(C.int(major), C.int(minor), C.int(patch)))
}
