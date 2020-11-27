package main

import (
	"errors"
	"flag"
	"fmt"
	"reflect"

	"github.com/stellentus/go-plc"
)

var (
	addr     = flag.String("address", "192.168.1.176", "Hostname or IP address of the PLC")
	plcDebug = flag.Int("plctagdebug", 0, "Debug level for libplctag's debug (0-5)")
)

func main() {
	flag.Parse()

	plc.SetLibplctagDebug(plc.LibplctagDebugLevel(*plcDebug))

	device, err := plc.NewDevice(*addr)
	panicIfError(err, "Could not create test PLC!")
	defer func() {
		err := device.Close()
		if err != nil {
			fmt.Println("Close was unsuccessful:", err.Error())
		}
	}()

	tags, err := device.GetAllTags()
	panicIfError(err, "Could not get PLC tags!")

	for _, tag := range tags {
		if !tag.CanBeInstantiated() {
			fmt.Printf("Cannot instantiate %v\n", tag)
			continue
		}

		newVal, err := tag.NewInstance()
		panicIfError(err, "Couldn't instantiate "+tag.Name())

		err = device.ReadTag(tag.Name(), newVal)
		switch {
		case errors.Is(err, plc.Pending):
			fmt.Printf("Tag %v is pending (%v)\n", tag, err)
		case errors.Is(err, plc.ErrBadRequest):
			fmt.Printf("Tag %v was a bad request (%v)\n", tag, err)
		case errors.Is(err, plc.ErrPlcConnection):
			fmt.Printf("Tag %v encountered a connection error (%v)\n", tag, err)
		case errors.Is(err, plc.ErrPlcInternal):
			fmt.Printf("Tag %v encountered an internal PLC error (%v)\n", tag, err)
		case err != nil:
			panicIfError(err, fmt.Sprintf("Couldn't read %v", tag))
		default:
			fmt.Printf("%s is %v (type: %s)\n", tag.Name(), reflect.ValueOf(newVal).Elem(), tag.TagType.String())
		}
	}
}

func panicIfError(err error, reason string) {
	if err != nil {
		panic("ERROR '" + err.Error() + "': " + reason)
	}
}
