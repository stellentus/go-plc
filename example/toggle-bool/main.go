package main

import (
	"flag"
	"fmt"

	"github.com/stellentus/go-plc"
)

var (
	addr     = flag.String("address", "192.168.1.176", "Hostname or IP address of the PLC")
	path     = flag.String("path", "1,0", "Path to the PLC at the provided host or IP")
	tagName  = flag.String("tagName", "DUMMY_BOOL_TAG", "Name of the boolean tag to toggle")
	plcDebug = flag.Int("plctagdebug", 0, "Debug level for libplctag's debug (0-5)")
)

func main() {
	flag.Parse()

	plc.SetLibplctagDebug(plc.LibplctagDebugLevel(*plcDebug))

	conf := map[string]string{
		"gateway": *addr,
		"path":    *path,
	}

	fmt.Println("Attempting test connection to", conf["gateway"], "using", *tagName)

	device, err := plc.NewDevice(conf)
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
	err = device.ReadTag(*tagName, &isOn)
	if err != nil {
		panic("ERROR: Unable to read the data because " + err.Error())
	}
	fmt.Printf("%s is %v\n", *tagName, isOn)

	// Toggle the bool state
	isOn = !isOn
	err = device.WriteTag(*tagName, isOn)
	if err != nil {
		panic("ERROR: Unable to write the data because " + err.Error())
	}

	// Confirm that it was toggled as expected
	var newIsOn bool
	err = device.ReadTag(*tagName, &newIsOn)
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
