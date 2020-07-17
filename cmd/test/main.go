package main

import (
	"fmt"

	"github.com/stellentus/go-plc"
)

func main() {
	connectionInfo := "protocol=ab_eip&gateway=192.168.29.121&path=1,0&cpu=LGX"
	tagName := "Enable_RampDown"
	timeout := 5000

	fmt.Println("Attempting test connection to", connectionInfo, "using", tagName)

	testPLC, err := plc.New(connectionInfo, timeout)
	if err != nil {
		panic("ERROR " + err.Error() + ": Could not create test PLC!")
	}
	defer func() {
		err := testPLC.Close()
		if err != nil {
			fmt.Println("Close was unsuccessful:", err.Error())
		}
	}()

	err = testPLC.StatusForTag(tagName)
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
	err = testPLC.ReadTag(tagName, &isOn)
	if err != nil {
		panic("ERROR: Unable to read the data because " + err.Error())
	}
	fmt.Printf("%s is %v\n", tagName, isOn)

	// Toggle the bool state
	isOn = !isOn
	err = testPLC.WriteTag(tagName, isOn)
	if err != nil {
		panic("ERROR: Unable to write the data because " + err.Error())
	}

	// Confirm that it was toggled as expected
	var newIsOn bool
	err = testPLC.ReadTag(tagName, &newIsOn)
	if err != nil {
		panic("ERROR: Unable to read the data because " + err.Error())
	}
	fmt.Printf("%s is %v\n", tagName, newIsOn)

	if isOn == newIsOn {
		fmt.Printf("SUCCESS! Bool switched from %v to %v as expected\n", !isOn, newIsOn)
	} else {
		fmt.Printf("FAILURE! value remained as %v\n", newIsOn)
	}
}
