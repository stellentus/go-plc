package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/stellentus/go-plc"
	"github.com/stellentus/go-plc/example"
)

var (
	plcAddr         = flag.String("plc-address", "192.168.1.176", "Hostname or IP address of the PLC")
	httpAddr        = flag.String("http", ":8784", "Port for http server to listen to")
	numWorkers      = flag.Int("workers", 1, "Number of worker threads talking to libplctag")
	refreshDuration = flag.Duration("refresh", time.Second, "Refresh period")
	useCache        = flag.Bool("usecache", false, "Cache values")
	plcDebug        = flag.Int("plctagdebug", 0, "Debug level for libplctag's debug (0-5)")
)

var knownTags = map[string]interface{}{
	"DUMMY_AQUA_DATA_0": uint16(0),
	"DUMMY_AQUA_DATA_1": uint16(0),
	"DUMMY_AQUA_DATA_2": uint16(0),
	"DUMMY_AQUA_DATA_3": uint16(0),
	"DUMMY_AQUA_DATA_4": uint16(0),
	"DUMMY_AQUA_DATA_5": uint16(0),
	"DUMMY_AQUA_DATA_6": uint16(0),
	"DUMMY_AQUA_DATA_7": uint16(0),
	"DUMMY_AQUA_DATA_8": uint16(0),
	"DUMMY_AQUA_DATA_9": uint16(0),
}

func main() {
	flag.Parse()

	if *useCache && *refreshDuration <= 0 {
		fmt.Println("Cannot use cache without a refresher")
		return
	}

	plc.SetLibplctagDebug(plc.LibplctagDebugLevel(*plcDebug))

	fmt.Printf("Initializing connection to %s\n", *plcAddr)

	device, err := plc.NewDevice(*plcAddr)
	panicIfError(err, "Could not create test PLC!")
	defer func() {
		err := device.Close()
		panicIfError(err, "Close was unsuccessful")
	}()

	// Wrap with debug
	rw := plc.ReadWriter(example.DebugPrinter{
		ReadPrefix: "READ",
		Reader:     device,
		Writer:     device,
		DebugFunc:  fmt.Printf,
	})

	if *numWorkers > 0 {
		fmt.Printf("Creating a pool of %d threads\n", *numWorkers)
		rw = plc.NewPooled(rw, *numWorkers)
	}

	// Now split into Reader and Writer chains.
	httpRW := struct {
		plc.Reader
		plc.Writer
	}{Reader: rw, Writer: rw}

	var cache *plc.Cache
	if *useCache {
		fmt.Printf("Creating a cache\n")
		cache = plc.NewCache(httpRW.Reader)
		httpRW.Reader = cache
	}

	var refresher plc.Reader
	if *refreshDuration > 0 {
		fmt.Printf("Creating a refresher to reload every %v\n", *refreshDuration)
		refresher = example.DebugPrinter{
			ReadPrefix: "REFRESH-START",
			Reader:     plc.NewRefresher(httpRW.Reader, *refreshDuration),
			DebugFunc:  fmt.Printf,
		}
		initializeRefresher(refresher)
	}

	// Can only use cache if there's also a refresher
	if *useCache && *refreshDuration > 0 {
		httpRW.Reader = example.DebugPrinter{
			ReadPrefix: "CACHE-READ",
			Reader:     cache.CacheReader(),
			DebugFunc:  fmt.Printf,
		}
	}

	http.Handle("/tags/raw", RawTagsHandler{httpRW, knownTags})
	fmt.Printf("Making PLC '%s' available at '%s'\n", *plcAddr, *httpAddr)
	err = http.ListenAndServe(*httpAddr, nil)
	panicIfError(err, "Could not start http server!")
}

func initializeRefresher(rd plc.Reader) error {
	if rd == nil {
		return nil // No refresher? Do nothing
	}
	for tag, v := range knownTags {
		val := copy(v)
		err := rd.ReadTag(tag, val)
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
