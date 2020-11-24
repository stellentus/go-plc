package example

import (
	"reflect"

	"github.com/stellentus/go-plc"
)

type DebugFunc func(string, ...interface{}) (int, error)

type DebugPrinter struct {
	ReadPrefix string
	plc.Reader
	plc.Writer
	DebugFunc
}

func (dp DebugPrinter) ReadTag(name string, value interface{}) error {
	err := dp.Reader.ReadTag(name, value)
	if err != nil {
		dp.DebugFunc("%s FAILURE for %s (%v)\n", dp.ReadPrefix, name, err)
	} else {
		dp.DebugFunc("%s: %s is %v\n", dp.ReadPrefix, name, reflect.ValueOf(value).Elem())
	}
	return err
}

func (dp DebugPrinter) WriteTag(name string, value interface{}) error {
	err := dp.Writer.WriteTag(name, value)
	if err != nil {
		dp.DebugFunc("Write FAILURE setting %s to %v (%v)\n", name, reflect.ValueOf(value), err)
	} else {
		dp.DebugFunc("Write: %s is %v\n", name, reflect.ValueOf(value))
	}
	return err
}
