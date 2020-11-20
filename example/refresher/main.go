package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/stellentus/go-plc"
	"github.com/stellentus/go-plc/example"
)

var addr = flag.String("address", "192.168.1.176", "Hostname or IP address of the PLC")
var path = flag.String("path", "1,0", "Path to the PLC at the provided host or IP")
var timeout = flag.Duration("timeout", 5*time.Second, "PLC communication timeout")
var numWorkers = flag.Int("workers", 1, "Number of worker threads talking to libplctag")
var refreshDuration = flag.Duration("refresh", time.Second, "Refresh period")
var tagName = flag.String("tagName", "DUMMY_AQUA_DATA_0[0]", "Name of the uint8 tag to read repeatedly")

// This command demonstrates setting up to read and write values from a plant.
func main() {
	flag.Parse()

	dev := example.NewCompositeDevice(*addr, *path, *timeout, example.Config{
		Workers:        *numWorkers,
		PrintReadDebug: true,
		DebugFunc:      fmt.Printf,
	})

	fmt.Printf("Creating a refresher to reload every %v\n", *refreshDuration)
	refresher := plc.NewRefresher(dev, *refreshDuration)

	val := uint8(0)
	refresher.ReadTag(*tagName, &val)
	time.Sleep(10 * time.Second)
}
