package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/stellentus/go-plc"
)

var (
	addr     = flag.String("address", "192.168.1.176", "Hostname or IP address of the PLC")
	plcDebug = flag.Int("plctagdebug", 0, "Debug level for libplctag's debug (0-5)")
	typeCSV  = flag.String("typeNamePath", "example/get-tags/tags.csv", "Path to a CSV containing 'tag-type-id,name'")
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
	if *typeCSV == "" {
		return nil // nothing to do
	}
	fi, err := os.Open(*typeCSV)
	if err != nil {
		return err
	}
	defer fi.Close()

	rd := csv.NewReader(fi)
	records, err := rd.ReadAll()
	if err != nil {
		return err
	}

	for i, row := range records {
		if len(row) != 2 {
			return fmt.Errorf("Could not parse CSV row %d: %v", i, row)
		}

		tt, err := strconv.ParseUint(row[0], 16, 16)
		if err != nil {
			return err
		}

		err = plc.RegisterTagTypeName(plc.TagType(tt), row[1])
		if err != nil {
			return err
		}
	}

	return nil
}

func panicIfError(err error, reason string) {
	if err != nil {
		panic("ERROR " + err.Error() + ": " + reason)
	}
}
