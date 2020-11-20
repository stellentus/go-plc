package example

import (
	"fmt"
	"reflect"
	"time"

	"github.com/stellentus/go-plc"
)

type Config struct {
	Workers int
}

func NewDebugPooledDevice(addr string, path string, timeout time.Duration, conf Config) plc.ReadWriteCloser {
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

	fmt.Printf("Creating a pool of %d threads\n", conf.Workers)
	pooled := plc.NewPooled(device, conf.Workers)

	debug := plc.ReaderFunc(func(name string, value interface{}) error {
		fmt.Printf("Read: %s is %v\n", name, reflect.ValueOf(value).Elem())
		return pooled.ReadTag(name, value)
	})

	return struct {
		plc.Reader
		plc.Writer
		plc.Closer
	}{
		Reader: debug,
		Writer: pooled,
		Closer: device,
	}
}
