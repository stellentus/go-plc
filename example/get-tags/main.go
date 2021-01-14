package main

import (
	"flag"
	"fmt"

	"github.com/stellentus/go-plc"
)

var (
	addr     = flag.String("address", "192.168.1.176", "Hostname or IP address of the PLC")
	plcDebug = flag.Int("plctagdebug", 0, "Debug level for libplctag's debug (0-5)")
)

func main() {
	flag.Parse()

	plc.SetLibplctagDebug(plc.LibplctagDebugLevel(*plcDebug))

	panicIfError(registerTagTypes(), "Couldn't register tag types")

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

	fmt.Println("Tags:", tags)

	programs, err := device.GetAllPrograms()
	panicIfError(err, "Could not get PLC programs!")

	fmt.Println("Programs:", programs)
}

func registerTagTypes() error {
	// Add the dummy tag type
	return plc.RegisterTagTypeName(0x2000, "dummy_tag_type")
}

func panicIfError(err error, reason string) {
	if err != nil {
		panic("ERROR " + err.Error() + ": " + reason)
	}
}
