package plc

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	ErrBadRequest    = errors.New("Invalid request")
	ErrPlcInternal   = errors.New("Internal PLC error")
	ErrPlcConnection = errors.New("PLC connection error")
	Pending          = errors.New("The PLC has not yet provided a result for the non-blocking request")
)

type ErrNonPointerRead struct {
	TagName string
	reflect.Kind
}

func (err ErrNonPointerRead) Error() string {
	return fmt.Sprintf("ReadTag expects a pointer type but got %v for tag '%s'", err.Kind, err.TagName)
}

func (err ErrNonPointerRead) Unwrap() error { return ErrBadRequest } // Even though we don't say "bad request", that's still this error's type
