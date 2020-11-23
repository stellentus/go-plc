package example

import (
	"reflect"

	"github.com/stellentus/go-plc"
)

type DebugFunc func(string, ...interface{}) (int, error)

type Config struct {
	// Workers creates a pool of workers if greater than 0.
	Workers int

	// PrintReadDebug creates a wrapper to print the value being read.
	PrintReadDebug bool

	// PrintWriteDebug creates a wrapper to print the value being written.
	PrintWriteDebug bool

	// DebugFunc prints debug.
	// If nil, nothing is printed.
	DebugFunc

	// DeviceConnection is the map used by plc.Device to initialize the connection.
	DeviceConnection map[string]string
}

type Device struct {
	plc.ReadWriteCloser
}

func NewDevice(conf Config) (Device, error) {
	if conf.DebugFunc == nil {
		conf.DebugFunc = doNothing
	}

	var err error
	dev := Device{}

	conf.DebugFunc("Initializing connection to %s\n", conf.DeviceConnection["gateway"])
	dev.ReadWriteCloser, err = plc.NewDevice(conf.DeviceConnection)
	if err != nil {
		return dev, err
	}

	if conf.PrintReadDebug {
		dev.ReadWriteCloser = PrintReadDebug(dev.ReadWriteCloser, conf.DebugFunc)
	}

	if conf.PrintWriteDebug {
		dev.ReadWriteCloser = PrintWriteDebug(dev.ReadWriteCloser, conf.DebugFunc)
	}

	if conf.Workers > 0 {
		conf.DebugFunc("Creating a pool of %d threads\n", conf.Workers)
		dev.ReadWriteCloser = plc.NewPooled(dev.ReadWriteCloser, conf.Workers).WithCloser(dev.ReadWriteCloser)
	}

	return dev, nil
}

func PrintReadDebug(rwc plc.ReadWriteCloser, rf DebugFunc) plc.ReadWriteCloser {
	read := rwc.ReadTag
	return plc.ReaderFunc(func(name string, value interface{}) error {
		err := read(name, value)
		if err != nil {
			rf("Read FAILURE for %s (%v)\n", name, err)
		} else {
			rf("Read: %s is %v\n", name, reflect.ValueOf(value).Elem())
		}
		return err
	}).WithWriteCloser(rwc)
}

func PrintWriteDebug(rwc plc.ReadWriteCloser, rf DebugFunc) plc.ReadWriteCloser {
	write := rwc.WriteTag
	return plc.WriterFunc(func(name string, value interface{}) error {
		err := write(name, value)
		if err != nil {
			rf("Write FAILURE setting %s to %v (%v)\n", name, reflect.ValueOf(value), err)
		} else {
			rf("Write: %s is %v\n", name, reflect.ValueOf(value))
		}
		return err
	}).WithReadCloser(rwc)
}

func doNothing(string, ...interface{}) (int, error) { return 0, nil }
