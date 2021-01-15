package plc

import (
	"fmt"
	"reflect"
	"strings"
)

// SplitReader splits reads of structs, arrays, and slices into separate reads of their components.
// It is important to note that ReadTag will attempt to read or write a slice or array up to its length.
// This might cause a PLC error if the operation goes out of bounds.
// It also means nothing will be read if a nil or empty slice is provided; this code cannot infer the length.
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
		return newErrNonPointerRead(name, v.Kind())
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
			fieldName, ok := getNameOfField(str, i, false)
			if !ok {
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
	case reflect.Array, reflect.Slice:
		arr := v.Elem()
		for idx := 0; idx < arr.Len(); idx++ {
			itemName := TagWithIndex(name, idx)
			item := arr.Index(idx)
			if !item.CanAddr() {
				err = fmt.Errorf("Cannot address %s", itemName)
				break
			}
			itemPointer := item.Addr().Interface()
			if item.Kind() == reflect.Ptr {
				// Since item actually is a pointer, we want its value instead.
				if item.IsNil() {
					// It's currently a nil pointer, so we need to allocate and set the value
					newVal := reflect.New(item.Type().Elem())
					item.Set(newVal)
				}
				itemPointer = item.Interface()
			}

			err = r.ReadTag(itemName, itemPointer)
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
			fieldName, ok := getNameOfField(str, i, true)
			if !ok {
				continue // Can't touch that
			}
			fieldName = name + "." + fieldName // add prefix
			fieldPointer := str.Field(i).Interface()

			err = sw.WriteTag(fieldName, fieldPointer)
			if err != nil {
				break
			}
		}
	case reflect.Array, reflect.Slice:
		arr := v
		for idx := 0; idx < arr.Len(); idx++ {
			itemName := TagWithIndex(name, idx)
			itemPointer := arr.Index(idx).Interface()
			err = sw.WriteTag(itemName, itemPointer)
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

// getNameOfField gets the name of field i in the provided struct str.
// The second return argument indicates whether it's ok to use the field. If false,
// the field should be skipped.
// It does not consider any struct fields other than 'omitempty', which indicates
// the field should be skipped if it's a zero value. This is only relevant if
// allowOmitEmpty is true.
func getNameOfField(str reflect.Value, i int, allowOmitEmpty bool) (string, bool) {
	field := str.Type().Field(i)
	plctag := field.Tag.Get("plctag")
	if plctag == "" {
		return field.Name, true
	}
	opts := strings.Split(plctag, ",")
	name := opts[0]
	switch name {
	case "-":
		return "", false // Ignore this field
	case "":
		name = field.Name // Use the field name as the name
	}

	if !allowOmitEmpty {
		return name, true
	}

	for _, opt := range opts[1:] {
		if opt == "omitempty" {
			return name, !str.Field(i).IsZero()
		}
		// else an unused option was included, which is odd but not an error
	}

	return name, true
}
