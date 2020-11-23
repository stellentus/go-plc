package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/stellentus/go-plc"
	"github.com/stellentus/go-plc/example"
)

var addr = flag.String("address", "192.168.1.176", "Hostname or IP address of the PLC")
var numWorkers = flag.Int("workers", 1, "Number of worker threads talking to libplctag")
var refreshDuration = flag.Duration("refresh", time.Second, "Refresh period")
var tagName = flag.String("tagName", "DUMMY_AQUA_DATA_0[0]", "Name of the uint8 tag to read repeatedly")

// This command demonstrates setting up to read and write values from a plant.
func main() {
	flag.Parse()

	dev, err := example.NewDevice(example.Config{
		Workers:          *numWorkers,
		PrintReadDebug:   true,
		PrintWriteDebug:  true,
		DebugFunc:        fmt.Printf,
		DeviceConnection: map[string]string{"gateway": *addr},
	})
	if err != nil {
		panic("ERROR " + err.Error() + ": Could not create test PLC!")
	}
	defer func() {
		err := dev.Close()
		if err != nil {
			panic("ERROR: Close was unsuccessful:" + err.Error())
		}
	}()

	fmt.Printf("Creating a refresher to reload every %v\n", *refreshDuration)
	refresher := plc.NewRefresher(dev, *refreshDuration)

	// Tell the refresher to begin reading
	val := uint8(0)
	refresher.ReadTag(*tagName, &val)

	// Just read for 2 seconds
	time.Sleep(2 * time.Second)

	// Now write a new value. It will still be read by the refresher.
	dev.WriteTag(*tagName, val+1)
	time.Sleep(2 * time.Second)

	// Now return to the original value.
	dev.WriteTag(*tagName, val)
	time.Sleep(2 * time.Second)
}
