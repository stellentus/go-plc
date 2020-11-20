package example

import (
	"fmt"
	"reflect"
	"time"

	"github.com/stellentus/go-plc"
)

type Config struct {
	// Workers creates a pool of workers if greater than 0.
	Workers int

	// PrintReadDebug creates a wrapper to print the value being read.
	PrintReadDebug bool

	// PrintWriteDebug creates a wrapper to print the value being written.
	PrintWriteDebug bool

	// DebugFunc prints debug about device construction.
	// If nil, nothing is printed.
	DebugFunc func(string, ...interface{}) (int, error)
}

func NewDevice(addr string, path string, timeout time.Duration, conf Config) (plc.ReadWriteCloser, error) {
	connectionInfo := fmt.Sprintf("protocol=ab_eip&gateway=%s&path=%s&cpu=controllogix", addr, path)

	if conf.DebugFunc == nil {
		conf.DebugFunc = doNothing
	}

	var rwc plc.ReadWriteCloser
	var err error

	conf.DebugFunc("Initializing connection to %s\n", connectionInfo)
	rwc, err = plc.NewDevice(connectionInfo, timeout)
	if err != nil {
		return nil, err
	}

	if conf.Workers > 0 {
		conf.DebugFunc("Creating a pool of %d threads\n", conf.Workers)
		rwc = plc.NewPooled(rwc, conf.Workers).WithCloser(rwc)
	}

	if conf.PrintReadDebug {
		rwc = PrintReadDebug(rwc)
	}

	if conf.PrintWriteDebug {
		rwc = PrintWriteDebug(rwc)
	}

	return rwc, nil
}

func PrintReadDebug(rwc plc.ReadWriteCloser) plc.ReadWriteCloser {
	read := rwc.ReadTag
	return plc.ReaderFunc(func(name string, value interface{}) error {
		err := read(name, value)
		if err != nil {
			fmt.Printf("Read FAILURE for %s (%v)\n", name, err)
		} else {
			fmt.Printf("Read: %s is %v\n", name, reflect.ValueOf(value).Elem())
		}
		return err
	}).WithWriteCloser(rwc)
}

func PrintWriteDebug(rwc plc.ReadWriteCloser) plc.ReadWriteCloser {
	write := rwc.WriteTag
	return plc.WriterFunc(func(name string, value interface{}) error {
		err := write(name, value)
		if err != nil {
			fmt.Printf("Write FAILURE setting %s to %v (%v)\n", name, reflect.ValueOf(value), err)
		} else {
			fmt.Printf("Write: %s is %v\n", name, reflect.ValueOf(value))
		}
		return err
	}).WithReadCloser(rwc)
}

func doNothing(string, ...interface{}) (int, error) { return 0, nil }
