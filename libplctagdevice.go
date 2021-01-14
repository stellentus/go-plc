package plc

/*
#include <stdlib.h>
#include "./libplctag.h"
*/
import "C"
import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
	"unsafe"
)

var (
	ErrBadRequest    = errors.New("Invalid request")
	ErrPlcInternal   = errors.New("Internal PLC error")
	ErrPlcConnection = errors.New("PLC connection error")
	Pending          = errors.New("The PLC has not yet provided a result for the non-blocking request")
)

// newLibplctagError attempts to classify PLC errors according to whether it's some issue in the user input
// or something internal to the PLC code (e.g. in go, libplctag, networking, or the PLC itsef).
func newLibplctagError(code C.int32_t) error {
	switch code {
	case C.PLCTAG_STATUS_OK:
		return nil

	// This isn't really an error, though our code shouldn't ever return it, so perhaps it should be ErrPlcInternal
	case C.PLCTAG_STATUS_PENDING:
		return fmt.Errorf("%w", Pending)

	// These are all bad requests
	case C.PLCTAG_ERR_BAD_CONFIG,
		C.PLCTAG_ERR_BAD_DEVICE, // trying to address something that doesn't exist
		C.PLCTAG_ERR_BAD_PARAM,  // this might indicate a problem with this go code
		C.PLCTAG_ERR_NOT_ALLOWED,
		C.PLCTAG_ERR_NOT_FOUND,
		C.PLCTAG_ERR_NO_DATA,
		C.PLCTAG_ERR_NO_MATCH,
		C.PLCTAG_ERR_OUT_OF_BOUNDS,
		C.PLCTAG_ERR_UNSUPPORTED:
		cstr := C.plc_tag_decode_error(C.int(code))
		return fmt.Errorf("%w: %s", ErrBadRequest, C.GoString(cstr))

	// These are all connection issues
	case C.PLCTAG_ERR_BAD_CONNECTION,
		C.PLCTAG_ERR_BAD_GATEWAY,
		C.PLCTAG_ERR_TIMEOUT,
		C.PLCTAG_ERR_PARTIAL:
		cstr := C.plc_tag_decode_error(C.int(code))
		return fmt.Errorf("%w: %s", ErrPlcConnection, C.GoString(cstr))

	// These are all internal errors
	case C.PLCTAG_ERR_ABORT, // This is likely a bug in this go code
		C.PLCTAG_ERR_BAD_DATA, // This could also be a connection issue
		C.PLCTAG_ERR_BAD_REPLY,
		C.PLCTAG_ERR_BAD_STATUS,
		C.PLCTAG_ERR_CLOSE,
		C.PLCTAG_ERR_CREATE,
		C.PLCTAG_ERR_DUPLICATE, // probably a libplctag error
		C.PLCTAG_ERR_ENCODE,    // probably a libplctag error, but could be an input issue
		C.PLCTAG_ERR_MUTEX_DESTROY,
		C.PLCTAG_ERR_MUTEX_INIT,
		C.PLCTAG_ERR_MUTEX_LOCK,
		C.PLCTAG_ERR_MUTEX_UNLOCK,
		C.PLCTAG_ERR_NOT_IMPLEMENTED,
		C.PLCTAG_ERR_NO_MEM,       // libplctag
		C.PLCTAG_ERR_NO_RESOURCES, // PLC
		C.PLCTAG_ERR_NULL_PTR,     // could also occur if an invalid handle is used, which would be a bug in go code
		C.PLCTAG_ERR_OPEN,
		C.PLCTAG_ERR_READ,
		C.PLCTAG_ERR_REMOTE_ERR,
		C.PLCTAG_ERR_THREAD_CREATE, // This might need special handling, as it indicates libplctag is in a *very* bad state
		C.PLCTAG_ERR_THREAD_JOIN,
		C.PLCTAG_ERR_TOO_LARGE, // more data returned than expected
		C.PLCTAG_ERR_TOO_SMALL,
		C.PLCTAG_ERR_WINSOCK,
		C.PLCTAG_ERR_WRITE,
		C.PLCTAG_ERR_BUSY:
		cstr := C.plc_tag_decode_error(C.int(code))
		return fmt.Errorf("%w: %s", ErrPlcInternal, C.GoString(cstr))

	default:
		cstr := C.plc_tag_decode_error(C.int(code))
		return fmt.Errorf("%w: Unclassified error (%s)", ErrPlcInternal, C.GoString(cstr))
	}
}

type LibplctagDebugLevel int

const (
	DebugNone = LibplctagDebugLevel(iota)
	DebugError
	DebugWarn
	DebugInfo
	DebugDetail
	DebugSpew
)

const SystemTagBit = 0x1000
const TagDimensionMask = 0x6000

func SetLibplctagDebug(level LibplctagDebugLevel) {
	C.plc_tag_set_debug_level(C.int(level))
}

// libplctagDevice is an instance of the rawDevice interface.
// It communicates with a PLC over the network by using the libplctag C library.
type libplctagDevice struct {
	conConf string
	ids     map[string]C.int32_t
	timeout C.int
}

var _ = rawDevice(&libplctagDevice{})  // Compiler makes sure this type is a rawDevice
var _ = ReadWriter(&libplctagDevice{}) // Compiler makes sure this type is a ReadWriter

// newLibplctagDevice creates a new libplctagDevice.
// The conConf string provides IP and other connection configuration (see libplctag for options).
// It is not thread safe.
func newLibplctagDevice(conConf string, timeout time.Duration) *libplctagDevice {
	return &libplctagDevice{
		conConf: conConf,
		ids:     make(map[string]C.int32_t),
		timeout: C.int(timeout.Milliseconds()),
	}
}

// Close should be called on the libplctagDevice to clean up its resources.
func (dev *libplctagDevice) Close() error {
	for _, id := range dev.ids {
		err := newLibplctagError(C.plc_tag_destroy(id))
		if err != nil {
			return err
		}
	}
	return nil
}

const (
	noOffset         = C.int(0)
	stringDataOffset = 4
	stringMaxLength  = 82 // Size according to libplctag. Seems like an underlying protocol thing.
)

func (dev *libplctagDevice) getID(tagName string) (C.int32_t, error) {
	id, ok := dev.ids[tagName]
	if ok {
		return id, nil
	}

	cattrib_str := C.CString(dev.conConf + "&name=" + tagName) // can also specify elem_size=1&elem_count=1
	defer C.free(unsafe.Pointer(cattrib_str))

	id = C.plc_tag_create(cattrib_str, dev.timeout)
	if id < 0 {
		return id, newLibplctagError(id)
	}
	dev.ids[tagName] = id
	return id, nil
}

// ReadTag reads the requested tag into the provided value.
// It is not thread safe. In a multi-threaded context, callers should ensure the appropriate
// portion of the tag tree is locked.
func (dev *libplctagDevice) ReadTag(name string, value interface{}) error {
	id, err := dev.getID(name)
	if err != nil {
		return fmt.Errorf("ReadTag: %w", err)
	}

	if err := newLibplctagError(C.plc_tag_read(id, dev.timeout)); err != nil {
		return fmt.Errorf("ReadTag: %w", err)
	}

	switch val := value.(type) {
	case *bool:
		result, err := getUint8(id, noOffset)
		if err != nil {
			return fmt.Errorf("ReadTag: %w", err)
		}
		*val = uint8(result) > 0
	case *uint8:
		result, err := getUint8(id, noOffset)
		if err != nil {
			return fmt.Errorf("ReadTag: %w", err)
		}
		*val = uint8(result)
	case *uint16:
		result, err := getUint16(id, noOffset)
		if err != nil {
			return fmt.Errorf("ReadTag: %w", err)
		}
		*val = uint16(result)
	case *uint32:
		result, err := getUint32(id, noOffset)
		if err != nil {
			return fmt.Errorf("ReadTag: %w", err)
		}
		*val = uint32(result)
	case *uint64:
		result, err := getUint64(id, noOffset)
		if err != nil {
			return fmt.Errorf("ReadTag: %w", err)
		}
		*val = uint64(result)
	case *int8:
		result, err := getInt8(id, noOffset)
		if err != nil {
			return fmt.Errorf("ReadTag: %w", err)
		}
		*val = int8(result)
	case *int16:
		result, err := getInt16(id, noOffset)
		if err != nil {
			return fmt.Errorf("ReadTag: %w", err)
		}
		*val = int16(result)
	case *int32:
		result, err := getInt32(id, noOffset)
		if err != nil {
			return fmt.Errorf("ReadTag: %w", err)
		}
		*val = int32(result)
	case *int64:
		result, err := getInt64(id, noOffset)
		if err != nil {
			return fmt.Errorf("ReadTag: %w", err)
		}
		*val = int64(result)
	case *float32:
		result, err := getFloat32(id, noOffset)
		if err != nil {
			return fmt.Errorf("ReadTag: %w", err)
		}
		*val = float32(result)
	case *float64:
		result, err := getFloat64(id, noOffset)
		if err != nil {
			return fmt.Errorf("ReadTag: %w", err)
		}
		*val = float64(result)
	default:
		return fmt.Errorf("ReadTag: %w: unknown type %T (%v)", ErrBadRequest, val, val)
	}

	return nil
}

// WriteTag writes the provided tag and value.
// It is not thread safe. In a multi-threaded context, callers should ensure the appropriate
// portion of the tag tree is locked.
func (dev *libplctagDevice) WriteTag(name string, value interface{}) error {
	id, err := dev.getID(name)
	if err != nil {
		return fmt.Errorf("WriteTag: %w", err)
	}

	switch val := value.(type) {
	case bool:
		b := C.uint8_t(0)
		if val {
			b = C.uint8_t(255)
		}
		err = newLibplctagError(C.plc_tag_set_uint8(id, noOffset, b))
	case uint8:
		err = newLibplctagError(C.plc_tag_set_uint8(id, noOffset, C.uint8_t(val)))
	case uint16:
		err = newLibplctagError(C.plc_tag_set_uint16(id, noOffset, C.uint16_t(val)))
	case uint32:
		err = newLibplctagError(C.plc_tag_set_uint32(id, noOffset, C.uint32_t(val)))
	case uint64:
		err = newLibplctagError(C.plc_tag_set_uint64(id, noOffset, C.uint64_t(val)))
	case int8:
		err = newLibplctagError(C.plc_tag_set_int8(id, noOffset, C.int8_t(val)))
	case int16:
		err = newLibplctagError(C.plc_tag_set_int16(id, noOffset, C.int16_t(val)))
	case int32:
		err = newLibplctagError(C.plc_tag_set_int32(id, noOffset, C.int32_t(val)))
	case int64:
		err = newLibplctagError(C.plc_tag_set_int64(id, noOffset, C.int64_t(val)))
	case float32:
		err = newLibplctagError(C.plc_tag_set_float32(id, noOffset, C.float(val)))
	case float64:
		err = newLibplctagError(C.plc_tag_set_float64(id, noOffset, C.double(val)))
	case string:
		// write the string length
		err = newLibplctagError(C.plc_tag_set_int32(id, noOffset, C.int32_t(len(val))))
		if err != nil {
			return fmt.Errorf("WriteTag: %w", err)
		}

		// copy the data
		for str_index := 0; str_index < stringMaxLength; str_index++ {
			byt := byte(0) // pad with zeroes after the string ended
			if str_index < len(val) {
				byt = val[str_index]
			}

			err = newLibplctagError(C.plc_tag_set_uint8(id, C.int(stringDataOffset+str_index), C.uint8_t(byt)))
			if err != nil {
				return fmt.Errorf("WriteTag: %w", err)
			}
		}
	default:
		err = fmt.Errorf("Type %T is unknown and can't be written (%v)", val, val)
	}
	if err != nil {
		return fmt.Errorf("WriteTag: %w", err)
	}

	// Read. If non-zero, value is true. Otherwise, it's false.
	if err := newLibplctagError(C.plc_tag_write(id, dev.timeout)); err != nil {
		return fmt.Errorf("WriteTag: %w", err)
	}

	return nil
}

func (dev *libplctagDevice) GetList(listName, prefix string) ([]Tag, []string, error) {
	if listName == "" {
		listName += "@tags"
	} else {
		listName += ".@tags"
	}

	id, err := dev.getID(listName)
	if err != nil {
		return nil, nil, fmt.Errorf("GetList: %w", err)
	}

	if err := newLibplctagError(C.plc_tag_read(id, dev.timeout)); err != nil {
		return nil, nil, fmt.Errorf("GetList: %w", err)
	}

	tags := []Tag{}
	programNames := []string{}

	offset := C.int(0)
	for {
		tag := Tag{}
		offset += 4

		tag.TagType = TagType(C.plc_tag_get_uint16(id, offset))
		offset += 2

		tag.elementSize = uint16(C.plc_tag_get_uint16(id, offset))
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
		} else if (tag.TagType & SystemTagBit) == SystemTagBit {
			// Do nothing for system tags
		} else {
			numDimensions := int((tag.TagType & TagDimensionMask) >> 13)
			if numDimensions != len(tag.dimensions) {
				return nil, nil, fmt.Errorf("GetList: %w: tag '%s' claims to have %d dimensions but has %d", ErrPlcInternal, tag.name, numDimensions, len(tag.dimensions))
			}

			tags = append(tags, tag)
		}

		if offset >= C.plc_tag_get_size(id) {
			break
		}
	}

	return tags, programNames, nil
}

func getBool(id C.int32_t, offset C.int) (bool, error) {
	result, err := getUint8(id, offset)
	return result > 0, err
}

func getUint8(id C.int32_t, offset C.int) (uint8, error) {
	result := uint8(C.plc_tag_get_uint8(id, offset))
	if result == math.MaxUint8 {
		// If libplctag returns this value, it might be an error, so check
		return result, newLibplctagError(C.plc_tag_status(id))
	}
	return result, nil
}

func getUint16(id C.int32_t, offset C.int) (uint16, error) {
	result := uint16(C.plc_tag_get_uint16(id, offset))
	if result == math.MaxUint16 {
		// If libplctag returns this value, it might be an error, so check
		return result, newLibplctagError(C.plc_tag_status(id))
	}
	return result, nil
}

func getUint32(id C.int32_t, offset C.int) (uint32, error) {
	result := uint32(C.plc_tag_get_uint32(id, offset))
	if result == math.MaxUint32 {
		// If libplctag returns this value, it might be an error, so check
		return result, newLibplctagError(C.plc_tag_status(id))
	}
	return result, nil
}

func getUint64(id C.int32_t, offset C.int) (uint64, error) {
	result := uint64(C.plc_tag_get_uint64(id, offset))
	if result == math.MaxUint64 {
		// If libplctag returns this value, it might be an error, so check
		return result, newLibplctagError(C.plc_tag_status(id))
	}
	return result, nil
}

func getInt8(id C.int32_t, offset C.int) (int8, error) {
	result := int8(C.plc_tag_get_int8(id, offset))
	if result == math.MinInt8 {
		// If libplctag returns this value, it might be an error, so check
		return result, newLibplctagError(C.plc_tag_status(id))
	}
	return result, nil
}

func getInt16(id C.int32_t, offset C.int) (int16, error) {
	result := int16(C.plc_tag_get_int16(id, offset))
	if result == math.MinInt16 {
		// If libplctag returns this value, it might be an error, so check
		return result, newLibplctagError(C.plc_tag_status(id))
	}
	return result, nil
}

func getInt32(id C.int32_t, offset C.int) (int32, error) {
	result := int32(C.plc_tag_get_int32(id, offset))
	if result == math.MinInt32 {
		// If libplctag returns this value, it might be an error, so check
		return result, newLibplctagError(C.plc_tag_status(id))
	}
	return result, nil
}

func getInt64(id C.int32_t, offset C.int) (int64, error) {
	result := int64(C.plc_tag_get_int64(id, offset))
	if result == math.MinInt64 {
		// If libplctag returns this value, it might be an error, so check
		return result, newLibplctagError(C.plc_tag_status(id))
	}
	return result, nil
}

func getFloat32(id C.int32_t, offset C.int) (float32, error) {
	result := float32(C.plc_tag_get_float32(id, offset))
	if result == math.SmallestNonzeroFloat32 {
		// If libplctag returns this value, it might be an error, so check
		return result, newLibplctagError(C.plc_tag_status(id))
	}
	return result, nil
}

func getFloat64(id C.int32_t, offset C.int) (float64, error) {
	result := float64(C.plc_tag_get_float64(id, offset))
	if result == math.SmallestNonzeroFloat64 {
		// If libplctag returns this value, it might be an error, so check
		return result, newLibplctagError(C.plc_tag_status(id))
	}
	return result, nil
}
