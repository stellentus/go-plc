package example

import (
	"fmt"
	"reflect"
	"time"

	"github.com/stellentus/go-plc"
)

type Config struct {
	Workers        int
	PrintReadDebug bool
}

func NewCompositeDevice(addr string, path string, timeout time.Duration, conf Config) plc.ReadWriteCloser {
	connectionInfo := fmt.Sprintf("protocol=ab_eip&gateway=%s&path=%s&cpu=controllogix", addr, path)

	fmt.Println("Initializing connection to", connectionInfo)
	device, err := plc.NewDevice(connectionInfo, timeout)
	if err != nil {
		panic("ERROR " + err.Error() + ": Could not create test PLC!")
	}
	defer func() {
		err := device.Close()
		if err != nil {
			fmt.Println("Close was unsuccessful:", err.Error())
		}
	}()

	rwc := plc.ReadWriteCloser(device)

	if conf.Workers > 0 {
		fmt.Printf("Creating a pool of %d threads\n", conf.Workers)
		rwc = plc.NewPooled(rwc, conf.Workers).WithCloser(rwc)
	}

	if conf.PrintReadDebug {
		rwc = PrintReadDebug(rwc)
	}

	return rwc
}

func PrintReadDebug(rwc plc.ReadWriteCloser) plc.ReadWriteCloser {
	read := rwc.ReadTag
	return plc.ReaderFunc(func(name string, value interface{}) error {
		fmt.Printf("Read: %s is %v\n", name, reflect.ValueOf(value).Elem())
		return read(name, value)
	}).WithWriteCloser(rwc)
}
