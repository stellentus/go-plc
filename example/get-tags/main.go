package main

import (
	"flag"
	"fmt"

	"github.com/stellentus/go-plc/physical"
)

var (
	addr     = flag.String("address", "192.168.1.176", "Hostname or IP address of the PLC")
	plcDebug = flag.Int("plctagdebug", 0, "Debug level for libplctag's debug (0-5)")
)

func main() {
	flag.Parse()

	physical.SetLibplctagDebug(physical.LibplctagDebugLevel(*plcDebug))

	device, err := physical.NewDevice(*addr)
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

func panicIfError(err error, reason string) {
	if err != nil {
		panic("ERROR " + err.Error() + ": " + reason)
	}
}
