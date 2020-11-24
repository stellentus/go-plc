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
		rw = DebugPrinter{
			ReadPrefix: "READ",
			Reader:     rw,
			Writer:     rw,
			DebugFunc:  conf.DebugFunc,
		}
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
			dev.cache = DebugPrinter{
				ReadPrefix: "CACHE-READ",
				Reader:     dev.cache,
				DebugFunc:  conf.DebugFunc,
			}
		}
	}

	if conf.RefresherDuration > 0 {
		fmt.Printf("Creating a refresher to reload every %v\n", conf.RefresherDuration)
		refresher := plc.NewRefresher(dev.Reader, conf.RefresherDuration)
		dev.Reader = refresher
		dev.refresher = refresher

		if conf.PrintIODebug {
			dev.refresher = DebugPrinter{
				ReadPrefix: "REFRESH-START",
				Reader:     dev.refresher,
				DebugFunc:  conf.DebugFunc,
			}
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

func doNothing(string, ...interface{}) (int, error) { return 0, nil }
