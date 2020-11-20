package main

import (
	"flag"
	"fmt"

	"github.com/stellentus/go-plc"
)

var addr = flag.String("address", "192.168.1.176", "Hostname or IP address of the PLC")
var path = flag.String("path", "1,0", "Path to the PLC at the provided host or IP")

func main() {
	flag.Parse()

	conf := map[string]string{
		"gateway": *addr,
		"path":    *path,
	}
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
