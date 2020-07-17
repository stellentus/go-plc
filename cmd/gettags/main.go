package main

import (
	"flag"
	"fmt"

	"github.com/stellentus/go-plc"
)

var addr = flag.String("address", "192.168.29.121", "Hostname or IP address of the PLC")
var path = flag.String("path", "1,0", "Path to the PLC at the provided host or IP")

const (
	REQ_VER_MAJOR = 2
	REQ_VER_MINOR = 1
	REQ_VER_PATCH = 0
)

func main() {
	if err := plc.CheckRequiredVersion(REQ_VER_MAJOR, REQ_VER_MINOR, REQ_VER_PATCH); err != nil {
		panic(fmt.Sprintf("Required PLC library version %d.%d.%d is not available", REQ_VER_MAJOR, REQ_VER_MINOR, REQ_VER_PATCH))
	}

	flag.Parse()

	connectionInfo := fmt.Sprintf("protocol=ab_eip&gateway=%s&path=%s&cpu=LGX", *addr, *path)
	timeout := 5000
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

	tags, err := testPLC.GetAllTags()
	if err != nil {
		panic("ERROR " + err.Error() + ": Could not get PLC tags!")
	}

	fmt.Println("Tags:", tags)

	programs, err := testPLC.GetAllPrograms()
	if err != nil {
		panic("ERROR " + err.Error() + ": Could not get PLC programs!")
	}

	fmt.Println("Programs:", programs)
}
