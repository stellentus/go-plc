package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/stellentus/go-plc"
	"github.com/stellentus/go-plc/example"
	"github.com/stellentus/go-plc/libplctag"
)

var (
	addr            = flag.String("address", "192.168.1.176", "Hostname or IP address of the PLC")
	numWorkers      = flag.Int("workers", 1, "Number of worker threads talking to libplctag")
	refreshDuration = flag.Duration("refresh", time.Second, "Refresh period")
	tagName         = flag.String("tagName", "DUMMY_TAG", "Name of the uint8 tag to read repeatedly")
	plcDebug        = flag.Int("plctagdebug", 0, "Debug level for libplctag's debug (0-5)")
)

// This command demonstrates setting up to read and write values from a plant.
func main() {
	flag.Parse()

	if *refreshDuration <= 0 {
		panic("Cannot test refresher with no duration")
	}

	libplctag.SetDebug(libplctag.DebugLevel(*plcDebug))

	fmt.Printf("Initializing connection to %s\n", *addr)

	device, err := libplctag.NewDevice(*addr)
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

	fmt.Printf("Creating a refresher to reload every %v\n", *refreshDuration)
	refresher := example.DebugPrinter{ // Wrap the referesher in debug
		ReadPrefix: "REFRESH-START",
		Reader:     plc.NewRefresher(rw, *refreshDuration),
		DebugFunc:  fmt.Printf,
	}

	// Tell the refresher to begin reading
	val := uint8(0)
	refresher.ReadTag(*tagName, &val)

	// Just read for 2 seconds
	time.Sleep(2 * time.Second)

	// Now write a new value. It will still be read by the refresher.
	rw.WriteTag(*tagName, val+1)
	time.Sleep(2 * time.Second)

	// Now return to the original value.
	rw.WriteTag(*tagName, val)
	time.Sleep(2 * time.Second)
}

func panicIfError(err error, reason string) {
	if err != nil {
		panic("ERROR " + err.Error() + ": " + reason)
	}
}
