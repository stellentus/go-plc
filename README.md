# go-plc

A wrapper for PLC communication in golang. Examples are provided in the [example](https://github.com/stellentus/go-plc/tree/master/example) directory. They expect to be run against the stub version of `libplctag`.

## Instructions

Try running `example/toggle-bool/main.go`:
* Install the [`libplctag`](https://github.com/libplctag/libplctag) dynamic library wherever your OS expects it. (Just follow the `libplctag` instructions to `make install`.)
* `go run example/toggle-bool/main.go` (or use `go build`)

## Running with the Stub

* Install the [stub version](https://github.com/dijkstracula/plcstub/) of the `libplctag` dynamic library.
* Build or run the code using the `stub` tag. e.g. `go run -tags stub example/toggle-bool/main.go`
