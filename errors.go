package plc

/*
#include "./libplctag.h"
*/
import "C"
import (
	"errors"
	"fmt"
)

var (
	ErrBadRequest  = errors.New("Invalid request")
	ErrPlcInternal = errors.New("Internal PLC error")
	Pending        = errors.New("The PLC has not yet provided a result for the non-blocking request")
)

func newLibplctagError(code C.int32_t) error {
	switch code {
	case C.PLCTAG_STATUS_OK:
		return nil
	case C.PLCTAG_STATUS_PENDING:
		return fmt.Errorf("%w", Pending)
	default:
		cstr := C.plc_tag_decode_error(C.int(code))
		return fmt.Errorf("%w: %s", ErrBadRequest, C.GoString(cstr))
	}
}
