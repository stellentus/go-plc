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

	err := error(nil)
	switch v.Elem().Kind() {
	case reflect.Struct:
		str := v.Elem()
		for i := 0; i < str.NumField(); i++ {
			if str.Type().Field(i).PkgPath != "" {
				continue // Type is not exported, so skip it
			}

			// Generate the name of the struct's field and recurse
			fieldName := getNameOfField(str, i)
			if fieldName == "" {
				continue // Can't touch that
			}
			fieldName = name + "." + fieldName // add prefix
			if !str.Field(i).CanAddr() {
				err = fmt.Errorf("Cannot address %s", fieldName)
				break
			}
			fieldPointer := str.Field(i).Addr().Interface()

			err = r.ReadTag(fieldName, fieldPointer)
			if err != nil {
				break
			}
		}
	default:
		// Just try with the underlying type
		err = r.Reader.ReadTag(name, value)
	}

	return err
}

func getNameOfField(str reflect.Value, i int) string {
	field := str.Type().Field(i)
	plctag := field.Tag.Get("plctag")
	if plctag == "" {
		return field.Name
	} else if plctag == "-" {
		return ""
	}
	return plctag
}
