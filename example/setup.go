package example

import (
	"fmt"
	"io"
	"reflect"
	"time"

	"github.com/stellentus/go-plc"
)

type DebugFunc func(string, ...interface{}) (int, error)

type Config struct {
	// Workers creates a pool of workers if greater than 0.
	Workers int

	// PrintIODebug creates a wrapper to print the values being read and written.
	PrintIODebug bool

	// DebugFunc prints debug.
	// If nil, nothing is printed.
	DebugFunc

	// DeviceConnection is the map used by plc.Device to initialize the connection.
	DeviceConnection map[string]string

	// UseCache creates a cache if true.
	UseCache bool

	// RefresherDuration creates a duration if not zero.
	RefresherDuration time.Duration
}

type Device struct {
	plc.Reader
	plc.Writer
	io.Closer
	cache     plc.Reader
	refresher plc.Reader
}

func NewDevice(conf Config) (Device, error) {
	if conf.DebugFunc == nil {
		conf.DebugFunc = doNothing
	}

	conf.DebugFunc("Initializing connection to %s\n", conf.DeviceConnection["gateway"])
	device, err := plc.NewDevice(conf.DeviceConnection)
	if err != nil {
		return Device{}, err
	}

	dev := Device{
		Closer: device,
	}

	// Now add wrappers that apply to the entire ReadWriter.
	rw := plc.ReadWriter(device)

	if conf.PrintIODebug {
		rw = newDebugReadWriter("READ", rw.ReadTag, rw.WriteTag, conf.DebugFunc)
	}

	if conf.Workers > 0 {
		conf.DebugFunc("Creating a pool of %d threads\n", conf.Workers)
		rw = plc.NewPooled(rw, conf.Workers)
	}

	// Now split into Reader and Writer chains.
	dev.Reader = rw
	dev.Writer = rw

	if conf.UseCache {
		conf.DebugFunc("Creating a cache\n")
		cache := plc.NewCache(dev.Reader)
		dev.Reader = cache

		dev.cache = cache.CacheReader()
		if conf.PrintIODebug {
			dev.cache = newPrintReaderFunc("CACHE-READ", dev.cache.ReadTag, conf.DebugFunc)
		}
	}

	if conf.RefresherDuration > 0 {
		fmt.Printf("Creating a refresher to reload every %v\n", conf.RefresherDuration)
		refresher := plc.NewRefresher(dev.Reader, conf.RefresherDuration)
		dev.Reader = refresher
		dev.refresher = refresher

		if conf.PrintIODebug {
			dev.refresher = newPrintReaderFunc("REFRESH-START", dev.refresher.ReadTag, conf.DebugFunc)
		}
	}

	return dev, nil
}

func (dev Device) Cache() plc.Reader {
	return dev.cache
}

func (dev Device) Refresher() plc.Reader {
	return dev.refresher
}

// ReaderFunc is a function that can be used as a Reader.
// It's the same pattern as http.HandlerFunc.
type ReaderFunc func(name string, value interface{}) error

func (f ReaderFunc) ReadTag(name string, value interface{}) error {
	return f(name, value)
}

// WriterFunc is a function that can be used as a Writer.
// It's the same pattern as http.HandlerFunc.
type WriterFunc func(name string, value interface{}) error

func (f WriterFunc) WriteTag(name string, value interface{}) error {
	return f(name, value)
}

func newPrintReaderFunc(prefix string, rd ReaderFunc, rf DebugFunc) ReaderFunc {
	return ReaderFunc(func(name string, value interface{}) error {
		err := rd(name, value)
		if err != nil {
			rf("%s FAILURE for %s (%v)\n", prefix, name, err)
		} else {
			rf("%s: %s is %v\n", prefix, name, reflect.ValueOf(value).Elem())
		}
		return err
	})
}

func newPrintWriterFunc(wr WriterFunc, rf DebugFunc) WriterFunc {
	return WriterFunc(func(name string, value interface{}) error {
		err := wr(name, value)
		if err != nil {
			rf("Write FAILURE setting %s to %v (%v)\n", name, reflect.ValueOf(value), err)
		} else {
			rf("Write: %s is %v\n", name, reflect.ValueOf(value))
		}
		return err
	})
}

func newDebugReadWriter(prefix string, rd ReaderFunc, wr WriterFunc, rf DebugFunc) plc.ReadWriter {
	return struct {
		plc.Reader
		plc.Writer
	}{
		newPrintReaderFunc(prefix, rd, rf),
		newPrintWriterFunc(wr, rf),
	}
}

func doNothing(string, ...interface{}) (int, error) { return 0, nil }
