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

	tags, err := device.GetAllTags()
	if err != nil {
		panic("ERROR " + err.Error() + ": Could not get PLC tags!")
	}

	fmt.Println("Tags:", tags)

	programs, err := device.GetAllPrograms()
	if err != nil {
		panic("ERROR " + err.Error() + ": Could not get PLC programs!")
	}

	fmt.Println("Programs:", programs)
}
