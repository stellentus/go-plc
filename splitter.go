package plc

import (
	"fmt"
	"reflect"
	"strings"
)

// TagPrefix is the prefix used for struct tags.
// Note the use of the word 'tag' in the prefix itself refers to PLC tags, not go tags.
const TagPrefix = "plctag"

// SplitReader splits reads of structs, arrays, and slices into separate reads of their components.
// It is important to note that ReadTag will attempt to read or write a slice or array up to its length.
// This might cause a PLC error if the operation goes out of bounds.
// It also means nothing will be read if a nil or empty slice is provided; this code cannot infer the length.
type SplitReader struct {
	Reader
	newAsyncer func(action) asyncer
}

var _ = Reader(SplitReader{}) // Compiler makes sure this type is a Reader

// NewSplitReader returns a SplitReader.
func NewSplitReader(rd Reader) SplitReader {
	return SplitReader{Reader: rd, newAsyncer: func(act action) asyncer { return newNotAsync(act) }}
}

// NewSplitReaderParallel returns a SplitReader which makes calls in parallel.
func NewSplitReaderParallel(rd Reader) SplitReader {
	return SplitReader{Reader: rd, newAsyncer: func(act action) asyncer { return newAsync(act) }}
}

func (rd SplitReader) ReadTag(name string, value interface{}) error {
	as := rd.newAsyncer(rd.Reader.ReadTag)
	rd.readTagAsync(name, value, as)
	return as.Wait()
}

func (rd SplitReader) readTagAsync(name string, value interface{}, as asyncer) {
	v := reflect.ValueOf(value)
	if v.Kind() != reflect.Ptr {
		as.AddError(ErrNonPointerRead{TagName: name, Kind: v.Kind()})
		return
	}

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
			if name != "" {
				fieldName = name + "." + fieldName // add prefix
			}
			field := str.Field(i)
			rd.readValue(fieldName, field, as)
		}
	case reflect.Array, reflect.Slice:
		arr := v.Elem()
		for idx := 0; idx < arr.Len(); idx++ {
			rd.readValue(TagWithIndex(name, idx), arr.Index(idx), as)
		}
	default:
		// Just try with the underlying type
		as.Add(name, value)
	}
}

func (rd SplitReader) readValue(name string, val reflect.Value, as asyncer) {
	if !val.CanAddr() {
		as.AddError(fmt.Errorf("Cannot address %s", name))
		return
	}

	valPointer := val.Addr().Interface()
	if val.Kind() == reflect.Ptr {
		// Since val actually is a pointer, we want its value instead.
		if val.IsNil() {
			// It's currently a nil pointer, so we need to allocate and set the value
			newVal := reflect.New(val.Type().Elem())
			val.Set(newVal)
		}
		valPointer = val.Interface()
	}

	rd.readTagAsync(name, valPointer, as)
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
			if name != "" {
				fieldName = name + "." + fieldName // add prefix
			}
			fieldPointer := str.Field(i).Interface()

			if err := sw.WriteTag(fieldName, fieldPointer); err != nil {
				return err
			}
		}
	case reflect.Array, reflect.Slice:
		arr := v
		for idx := 0; idx < arr.Len(); idx++ {
			itemName := TagWithIndex(name, idx)
			itemPointer := arr.Index(idx).Interface()
			if err := sw.WriteTag(itemName, itemPointer); err != nil {
				return err
			}
		}
	default:
		// Just try with the underlying type
		return sw.Writer.WriteTag(name, v.Interface())
	}

	return nil
}

// getNameOfField gets the name of field i in the provided struct str.
// The second return argument indicates whether it's ok to use the field. If false,
// the field should be skipped.
// It does not consider any struct fields other than 'omitempty', which indicates
// the field should be skipped if it's a zero value. This is only relevant if
// allowOmitEmpty is true.
func getNameOfField(str reflect.Value, i int, allowOmitEmpty bool) (string, bool) {
	field := str.Type().Field(i)
	plctag := field.Tag.Get(TagPrefix)
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
