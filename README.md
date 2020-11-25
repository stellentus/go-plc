# go-plc

A wrapper for PLC communication in golang. Examples are provided in the [example](https://github.com/stellentus/go-plc/tree/master/example) directory. They expect to be run against the stub version of `libplctag`.

## Instructions

Try running `example/toggle-bool/main.go`:
* Compile or download `libplctag.a` for your platform. Place it at the root of this project.
* `go run example/toggle-bool/main.go` (or use `go build`)

## Running with the Stub

* Download and build the [stub version](https://github.com/dijkstracula/plcstub/) of `libplctag`
* Rename it to `libplctag_stub.a` and place it at the root of this project.
* Build or run the code using the `stub` tag. e.g. `go run -tags stub example/toggle-bool/main.go`
