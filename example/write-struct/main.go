package main

import (
	"flag"
	"fmt"

	"github.com/stellentus/go-plc"
	"github.com/stellentus/go-plc/physical"
)

var (
	addr     = flag.String("address", "192.168.1.176", "Hostname or IP address of the PLC")
	plcDebug = flag.Int("plctagdebug", 0, "Debug level for libplctag's debug (0-5)")
)

type DemoTags struct {
	Speed    int64
	Duration float32 `plctag:"THE_DURATION"`
	Level    int32   `plctag:",omitempty"`
}

// This command demonstrates writing a struct.
// It will write the struct, then read each value individually.
func main() {
	flag.Parse()

	physical.SetLibplctagDebug(physical.LibplctagDebugLevel(*plcDebug))

	fmt.Printf("Initializing connection to %s\n", *addr)

	device, err := physical.NewDevice(*addr)
	panicIfError(err, "Could not create test PLC!")
	defer func() {
		err := device.Close()
		panicIfError(err, "Close was unsuccessful")
	}()

	// First write the Level to show it's not overwritten later
	originalLevel := int32(13)
	err = device.WriteTag("DUMMY_STRUCT.Level", originalLevel)
	panicIfError(err, "Couldn't write Level")

	demoVal := DemoTags{
		Speed:    7738,
		Duration: 3.14,
		Level:    0, // Since this is 0 and is 'omitempty', it won't overwrite the existing value
	}

	// Wrap the device in a SplitWriter to allow a struct to be written
	err = plc.NewSplitWriter(device).WriteTag("DUMMY_STRUCT", demoVal)
	panicIfError(err, "Couldn't read dummy struct")

	// Demonstrate reading each of the three struct values independently
	var speed int64
	err = device.ReadTag("DUMMY_STRUCT.Speed", &speed)
	panicIfError(err, "Couldn't read Speed")
	fmt.Printf("Read speed as %d (expecting %d)\n", speed, demoVal.Speed)

	var duration float32
	err = device.ReadTag("DUMMY_STRUCT.THE_DURATION", &duration)
	panicIfError(err, "Couldn't read THE_DURATION")
	fmt.Printf("Read duration as %f (expecting %f)\n", duration, demoVal.Duration)

	var level int64
	err = device.ReadTag("DUMMY_STRUCT.Level", &level)
	panicIfError(err, "Couldn't read Level")
	fmt.Printf("Read level as %d (expecting %d)\n", level, originalLevel)

	// Now read everything by wrapping it with a SplitReader
	var readAll DemoTags
	err = plc.NewSplitReader(device).ReadTag("DUMMY_STRUCT", &readAll)
	panicIfError(err, "Couldn't read entire struct")
	demoVal.Level = originalLevel // Update to print the expectation
	fmt.Printf("Read struct as %v (expecting %v)\n", readAll, demoVal)
}

func panicIfError(err error, reason string) {
	if err != nil {
		panic("ERROR " + err.Error() + ": " + reason)
	}
}
