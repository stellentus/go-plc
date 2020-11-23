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
// Most of the main function is identical to example/refresher. Only the last few lines are different.
func main() {
	flag.Parse()

	dev, err := example.NewDevice(example.Config{
		Workers:          *numWorkers,
		PrintReadDebug:   true,
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

	fmt.Printf("Creating a cache\n")
	cache := plc.NewCache(dev)

	// Get the first read
	val := uint8(0)
	original := uint8(0)
	cache.ReadTag(*tagName, &original)
	cache.ReadCachedTag(*tagName, &val)
	fmt.Printf("Cached: %s is %v\n", *tagName, val)

	// Now write a new value, but re-read from the cache
	fmt.Println("Writing", val+1)
	dev.WriteTag(*tagName, val+1)
	time.Sleep(200 * time.Millisecond) // Arbitrary time to make sure the write completed
	cache.ReadCachedTag(*tagName, &val)
	fmt.Printf("Cached: %s is %v\n", *tagName, val)

	// Now read the value and show the updated cache
	cache.ReadTag(*tagName, &val)
	cache.ReadCachedTag(*tagName, &val)
	fmt.Printf("Cached: %s is %v\n", *tagName, val)

	// Now return to the original value.
	fmt.Println("Writing back to", original)
	dev.WriteTag(*tagName, original)
}
