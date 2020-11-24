package main

import (
	"flag"
	"fmt"

	"github.com/stellentus/go-plc"
)

var (
	addr     = flag.String("address", "192.168.1.176", "Hostname or IP address of the PLC")
	tagName  = flag.String("tagName", "DUMMY_AQUA_DATA_0[0]", "Name of the boolean tag to toggle")
	index    = flag.Int("index", -1, "Array index to access, or -1 if not an array")
	plcDebug = flag.Int("plctagdebug", 0, "Debug level for libplctag's debug (0-5)")
)

func main() {
	flag.Parse()

	plc.SetLibplctagDebug(plc.LibplctagDebugLevel(*plcDebug))

	fmt.Println("Attempting test connection to", *addr, "using", *tagName)

	device, err := plc.NewDevice(*addr)
	if err != nil {
		panic("ERROR " + err.Error() + ": Could not create test PLC!")
	}
	defer func() {
		err := device.Close()
		if err != nil {
			fmt.Println("Close was unsuccessful:", err.Error())
		}
	}()

	// Read. If non-zero, value is true. Otherwise, it's false.
	var isOn bool
	if *index >= 0 {
		err = device.ReadTag(plc.TagWithIndex(*tagName, *index), &isOn)
	} else {
		err = device.ReadTag(*tagName, &isOn)
	}
	if err != nil {
		panic("ERROR: Unable to read the data because " + err.Error())
	}
	fmt.Printf("%s is %v\n", *tagName, isOn)

	// Toggle the bool state
	isOn = !isOn
	if *index >= 0 {
		err = device.WriteTag(plc.TagWithIndex(*tagName, *index), isOn)
	} else {
		err = device.WriteTag(*tagName, isOn)
	}
	if err != nil {
		panic("ERROR: Unable to write the data because " + err.Error())
	}

	// Confirm that it was toggled as expected
	var newIsOn bool
	if *index >= 0 {
		err = device.ReadTag(plc.TagWithIndex(*tagName, *index), &newIsOn)
	} else {
		err = device.ReadTag(*tagName, &newIsOn)
	}
	if err != nil {
		panic("ERROR: Unable to read the data because " + err.Error())
	}
	fmt.Printf("%s is %v\n", *tagName, newIsOn)

	if isOn == newIsOn {
		fmt.Printf("SUCCESS! Bool switched from %v to %v as expected\n", !isOn, newIsOn)
	} else {
		fmt.Printf("FAILURE! value remained as %v\n", newIsOn)
	}
}
