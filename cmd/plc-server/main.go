package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/stellentus/go-plc"
)

var (
	plcAddr  = flag.String("plc-address", "192.168.1.176", "Hostname or IP address of the PLC")
	path     = flag.String("path", "1,0", "Path to the PLC at the provided host or IP")
	httpAddr = flag.String("http", ":8784", "Port for http server to listen to")
)

var knownTags = map[string]interface{}{
	"Program:ZW.B801.INISCP":      uint32(0),
	"Program:ZW.B801.MAXRANGE_SP": uint32(0),
	"Program:ZW.B801.SPEED_SP_1":  uint32(0),
	"Program:ZW.B801.SPEED_SP_2":  uint32(0),
}

func main() {
	flag.Parse()

	connectionInfo := fmt.Sprintf("protocol=ab_eip&gateway=%s&path=%s&cpu=controllogix", *plcAddr, *path)
	timeout := 5 * time.Second
	device, err := plc.NewDevice(connectionInfo, timeout)
	panicIfError(err, "Could not create test PLC!")
	defer func() {
		err := device.Close()
		if err != nil {
			fmt.Println("Close was unsuccessful:", err.Error())
		}
	}()

	http.Handle("/tags/raw", RawTagsHandler{device, knownTags})
	fmt.Printf("Making PLC '%s' available at '%s'\n", *plcAddr, *httpAddr)
	err = http.ListenAndServe(*httpAddr, nil)
	panicIfError(err, "Could not start http server!")
}

func panicIfError(err error, reason string) {
	if err != nil {
		panic("ERROR " + err.Error() + ": " + reason)
	}
}
