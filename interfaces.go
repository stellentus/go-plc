package plc

import (
	"fmt"
	"reflect"
)

type ReadWriter interface {
	Reader
	Writer
}

// Reader is the interface that wraps the basic ReadTag method.
type Reader interface {
	// ReadTag reads the requested tag into the provided value.
	ReadTag(name string, value interface{}) error
}

// Writer is the interface that wraps the basic WriteTag method.
type Writer interface {
	// WriteTag writes the provided tag and value.
	WriteTag(name string, value interface{}) error
}

// Closer is the interface that wraps the basic Close method.
//
// The behavior of Close after the first call is undefined.
// Specific implementations may document their own behavior.
type Closer interface {
	Close() error
}

// FakeReadWriter is provided as an example ReadWriter implementation and for use in tests.
type FakeReadWriter map[string]interface{}

func (df FakeReadWriter) ReadTag(name string, value interface{}) error {
	v, ok := df[name]
	if !ok {
		return fmt.Errorf("FakeReadWriter does not contain '%s'", name)
	}

	in := reflect.ValueOf(v)
	out := reflect.Indirect(reflect.ValueOf(value))

	switch {
	case !out.CanSet():
		return fmt.Errorf("FakeReadWriter for '%s', cannot set %s", name, out.Type().Name())
	case out.Kind() != in.Kind():
		return fmt.Errorf("FakeReadWriter for '%s', cannot set %s to %s (%v)", name, out.Type().Name(), in.Type().Name(), v)
	}

	out.Set(in)
	return nil
}

func (df FakeReadWriter) WriteTag(name string, value interface{}) error {
	df[name] = value
	return nil
}
