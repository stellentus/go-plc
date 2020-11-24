package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/stellentus/go-plc"
	"github.com/stellentus/go-plc/example"
)

var (
	addr       = flag.String("address", "192.168.1.176", "Hostname or IP address of the PLC")
	numWorkers = flag.Int("workers", 1, "Number of worker threads talking to libplctag")
	tagName    = flag.String("tagName", "DUMMY_TAG", "Name of the uint8 tag to read repeatedly")
	plcDebug   = flag.Int("plctagdebug", 0, "Debug level for libplctag's debug (0-5)")
)

// This command demonstrates setting up to read and write values from a plant.
func main() {
	flag.Parse()

	plc.SetLibplctagDebug(plc.LibplctagDebugLevel(*plcDebug))

	fmt.Printf("Initializing connection to %s\n", *addr)

	device, err := plc.NewDevice(*addr)
	panicIfError(err, "Could not create test PLC!")
	defer func() {
		err := device.Close()
		panicIfError(err, "Close was unsuccessful")
	}()

	// Wrap with debug
	rw := plc.ReadWriter(example.DebugPrinter{
		ReadPrefix: "READ",
		Reader:     device,
		Writer:     device,
		DebugFunc:  fmt.Printf,
	})

	if *numWorkers > 0 {
		fmt.Printf("Creating a pool of %d threads\n", *numWorkers)
		rw = plc.NewPooled(rw, *numWorkers)
	}

	fmt.Printf("Creating a cache\n")
	directReader := plc.NewCache(rw) // Reads the device directly

	// cacheReader only reads out of the cache
	cacheReader := example.DebugPrinter{
		ReadPrefix: "CACHE-READ",
		Reader:     directReader.CacheReader(),
		DebugFunc:  fmt.Printf,
	}

	// Get the first read
	val := uint8(0)
	original := uint8(0)
	directReader.ReadTag(*tagName, &original)
	cacheReader.ReadTag(*tagName, &val)

	// Now write a new value, but re-read from the cache
	rw.WriteTag(*tagName, val+1)
	time.Sleep(200 * time.Millisecond) // Arbitrary time to make sure the write completed
	cacheReader.ReadTag(*tagName, &val)

	// Now read the value and show the updated cache
	directReader.ReadTag(*tagName, &val)
	cacheReader.ReadTag(*tagName, &val)

	// Now return to the original value.
	rw.WriteTag(*tagName, original)
}

func panicIfError(err error, reason string) {
	if err != nil {
		panic("ERROR " + err.Error() + ": " + reason)
	}
}
