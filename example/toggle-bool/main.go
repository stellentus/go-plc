package main

import (
	"flag"
	"fmt"

	"github.com/stellentus/go-plc/physical"
)

var (
	addr     = flag.String("address", "192.168.1.176", "Hostname or IP address of the PLC")
	tagName  = flag.String("tagName", "DUMMY_BOOL_TAG", "Name of the boolean tag to toggle")
	plcDebug = flag.Int("plctagdebug", 0, "Debug level for libplctag's debug (0-5)")
)

func main() {
	flag.Parse()

	physical.SetLibplctagDebug(physical.LibplctagDebugLevel(*plcDebug))

	fmt.Println("Attempting test connection to", *addr, "using", *tagName)

	device, err := physical.NewDevice(*addr)
	panicIfError(err, "Could not create test PLC!")
	defer func() {
		err := device.Close()
		if err != nil {
			fmt.Println("Close was unsuccessful:", err.Error())
		}
	}()

	// Read. If non-zero, value is true. Otherwise, it's false.
	var isOn bool
	err = device.ReadTag(*tagName, &isOn)
	panicIfError(err, "Unable to read the data")
	fmt.Printf("%s is %v\n", *tagName, isOn)

	// Toggle the bool state
	isOn = !isOn
	err = device.WriteTag(*tagName, isOn)
	panicIfError(err, "Unable to write the data")

	// Confirm that it was toggled as expected
	var newIsOn bool
	err = device.ReadTag(*tagName, &newIsOn)
	panicIfError(err, "Unable to read the data")
	fmt.Printf("%s is %v\n", *tagName, newIsOn)

	if isOn == newIsOn {
		fmt.Printf("SUCCESS! Bool switched from %v to %v as expected\n", !isOn, newIsOn)
	} else {
		fmt.Printf("FAILURE! value remained as %v\n", newIsOn)
	}
}

func panicIfError(err error, reason string) {
	if err != nil {
		panic("ERROR " + err.Error() + ": " + reason)
	}
}
