package main

import (
	"flag"
	"fmt"
	"strconv"

	"github.com/stellentus/go-plc"
	"github.com/stellentus/go-plc/physical"
)

var (
	addr     = flag.String("address", "192.168.1.176", "Hostname or IP address of the PLC")
	tagName  = flag.String("tagName", "DUMMY_ARRAY_TAG", "Name of the array tag to read")
	index    = flag.Int("index", 0, "Array index to access")
	plcDebug = flag.Int("plctagdebug", 0, "Debug level for libplctag's debug (0-5)")
)

func main() {
	flag.Parse()
	if *index < 0 {
		panic("Cannot access negative array index: " + strconv.Itoa(*index))
	}

	physical.SetLibplctagDebug(physical.LibplctagDebugLevel(*plcDebug))

	fmt.Printf("Initializing connection to %s using %s\n", *addr, *tagName)

	device, err := physical.NewDevice(*addr)
	panicIfError(err, "Could not create test PLC!")
	defer func() {
		err := device.Close()
		if err != nil {
			fmt.Println("Close was unsuccessful:", err.Error())
		}
	}()

	var tagValue uint32
	err = device.ReadTag(plc.TagWithIndex(*tagName, *index), &tagValue)
	panicIfError(err, "Unable to read the data")
	fmt.Printf("%s[%d] is %v\n", *tagName, *index, tagValue)
}

func panicIfError(err error, reason string) {
	if err != nil {
		panic("ERROR " + err.Error() + ": " + reason)
	}
}
