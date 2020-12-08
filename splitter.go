package plc

import (
	"fmt"
	"reflect"
	"strings"
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
			fieldName, _ := getNameOfField(str, i)
			if fieldName == "" {
				continue // Can't touch that
			}
			fieldName = name + "." + fieldName // add prefix
			field := str.Field(i)
			if !field.CanAddr() {
				err = fmt.Errorf("Cannot address %s", fieldName)
				break
			}

			fieldPointer := field.Addr().Interface()
			if field.Kind() == reflect.Ptr {
				// Since field actually is a pointer, we want its value instead.
				if field.IsNil() {
					// It's currently a nil pointer, so we need to allocate and set the value
					newVal := reflect.New(field.Type().Elem())
					field.Set(newVal)
				}
				fieldPointer = field.Interface()
			}

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

// SplitWriter splits writes of structs and arrays into separate writes of their components.
type SplitWriter struct {
	Writer
}

var _ = Writer(SplitWriter{}) // Compiler makes sure this type is a Writer

// NewSplitWriter returns a SplitWriter.
func NewSplitWriter(wr Writer) SplitWriter {
	return SplitWriter{wr}
}

func (sw SplitWriter) WriteTag(name string, value interface{}) error {
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr {
		v = v.Elem() // Naturally use what the pointer is pointing to (but only do so once)
	}

	err := error(nil)
	switch v.Kind() {
	case reflect.Struct:
		str := v
		for i := 0; i < str.NumField(); i++ {
			if str.Type().Field(i).PkgPath != "" {
				continue // Type is not exported, so skip it
			}

			// Generate the name of the struct's field and recurse
			fieldName, omitempty := getNameOfField(str, i)
			if fieldName == "" {
				continue // Can't touch that
			}
			if omitempty && str.Field(i).IsZero() {
				continue // Skip zero values
			}
			fieldName = name + "." + fieldName // add prefix
			fieldPointer := str.Field(i).Interface()

			err = sw.WriteTag(fieldName, fieldPointer)
			if err != nil {
				break
			}
		}
	default:
		// Just try with the underlying type
		err = sw.Writer.WriteTag(name, v.Interface())
	}

	return err
}

func getNameOfField(str reflect.Value, i int) (string, bool) {
	omitEmpty := false

	field := str.Type().Field(i)
	plctag := field.Tag.Get("plctag")
	if plctag == "" {
		return field.Name, omitEmpty
	}
	opts := strings.Split(plctag, ",")
	switch opts[0] {
	case "-":
		return "", omitEmpty // Ignore this field
	case "":
		opts[0] = field.Name // Use the field name as the name
	}

	for _, opt := range opts[1:] {
		if opt == "omitempty" {
			omitEmpty = true
		} // else an unused option was included
	}

	return opts[0], omitEmpty
}
