package main

import (
	"flag"
	"fmt"
	"reflect"
	"time"

	"github.com/stellentus/go-plc"
)

var addr = flag.String("address", "192.168.1.176", "Hostname or IP address of the PLC")
var path = flag.String("path", "1,0", "Path to the PLC at the provided host or IP")
var numWorkers = flag.Int("workers", 1, "Number of worker threads talking to libplctag")
var timeout = flag.Duration("timeout", 5*time.Second, "PLC communication timeout")
var refreshDuration = flag.Duration("refresh", time.Second, "Refresh period")
var tagName = flag.String("tagName", "DUMMY_AQUA_DATA_0[0]", "Name of the uint8 tag to read repeatedly")

// This command demonstrates setting up to read and write values from a plant.
func main() {
	flag.Parse()

	refresher, _ := newPlant()

	val := uint8(0)
	refresher.ReadTag(*tagName, &val)
	time.Sleep(10 * time.Second)
}

func newPlant() (refresher plc.Reader, plant plc.ReadWriter) {
	connectionInfo := fmt.Sprintf("protocol=ab_eip&gateway=%s&path=%s&cpu=controllogix", *addr, *path)

	fmt.Println("Initializing connection to", connectionInfo)
	device, err := plc.NewDevice(connectionInfo, *timeout)
	if err != nil {
		panic("ERROR " + err.Error() + ": Could not create test PLC!")
	}
	// WARNING device.Close() should be called

	fmt.Printf("Creating a pool of %d threads\n", *numWorkers)
	pooled := plc.NewPooled(device, *numWorkers)

	debug := ReaderFunc(func(name string, value interface{}) error {
		fmt.Printf("Read: %s is %v\n", name, reflect.ValueOf(value).Elem())
		return pooled.ReadTag(name, value)
	})

	fmt.Printf("Creating a refresher to reload every %v\n", *refreshDuration)
	refresher = plc.NewRefresher(debug, *refreshDuration)

	return refresher, pooled
}

// ReaderFunc is a function that can be used as a Reader.
// It's the same pattern as http.HandlerFunc.
// Maybe this should eventually be moved into the package if it seems useful.
type ReaderFunc func(name string, value interface{}) error

func (rt ReaderFunc) ReadTag(name string, value interface{}) error {
	return rt(name, value)
}
