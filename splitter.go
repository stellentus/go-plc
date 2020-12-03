package plc

import (
	"fmt"
	"reflect"
)

// SplitReader splits reads of structs and arrays into separate reads of their components.
type SplitReader struct {
	Reader
}

var _ = Reader(SplitReader{}) // Compiler makes sure this type is a Reader

// NewSplitReader returns a SplitReader.
func NewSplitReader(rd Reader) SplitReader {
	return SplitReader{rd}
}

func (r SplitReader) ReadTag(name string, value interface{}) error {
	v := reflect.ValueOf(value)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("ReadTag expects a pointer type but got %v", v.Kind())
	}

	var err error
	switch v.Elem().Kind() {
	default:
		// Just try with the underlying type
		err = r.Reader.ReadTag(name, value)
	}

	return err
}
