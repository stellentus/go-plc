# go-plc

A wrapper for PLC communication in golang.

## Instructions

To run `cmd/test/main.go`:
* Compile or download `libplctag.a` for your platform. Place it at the root of this project.
* `go run cmd/test/main.go` (or use `go build`)

## Running with the Stub

* Download the [stub version](https://github.com/dijkstracula/plcstub/) of `libplctag`
* Rename it to `libplctag_stub.a` and place it at the root of this project.
* Build or run the code using the `stub` tag: `go run -tags stub cmd/test/main.go`
