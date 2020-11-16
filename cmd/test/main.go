package main

import (
	"flag"
	"fmt"

	"github.com/stellentus/go-plc"
)

var addr = flag.String("address", "192.168.1.176", "Hostname or IP address of the PLC")
var path = flag.String("path", "1,0", "Path to the PLC at the provided host or IP")
var tagName = flag.String("tagName", "Enable_RampDown", "Name of the boolean tag to toggle")
var index = flag.Int("index", -1, "Array index to access, or -1 if not an array")

func main() {
	flag.Parse()

	connectionInfo := fmt.Sprintf("protocol=ab_eip&gateway=%s&path=%s&cpu=LGX", *addr, *path)
	timeout := 5000

	fmt.Println("Attempting test connection to", connectionInfo, "using", *tagName)

	device, err := plc.NewDevice(connectionInfo, timeout)
	if err != nil {
		panic("ERROR " + err.Error() + ": Could not create test PLC!")
	}
	defer func() {
		err := device.Close()
		if err != nil {
			fmt.Println("Close was unsuccessful:", err.Error())
		}
	}()

	err = device.StatusForTag(*tagName)
	if err != nil {
		if _, ok := err.(plc.Pending); ok {
			panic("ERROR: PLC is not ready to communicate yet.")
		} else {
			panic("ERROR " + err.Error() + ": Error setting up tag internal state.")
		}
	}

	fmt.Println("Connected successfully")

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
