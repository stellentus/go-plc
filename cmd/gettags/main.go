package main

import (
	"fmt"

	"github.com/stellentus/go-plc"
)

const (
	REQ_VER_MAJOR = 2
	REQ_VER_MINOR = 1
	REQ_VER_PATCH = 0
)

func main() {
	if err := plc.CheckRequiredVersion(REQ_VER_MAJOR, REQ_VER_MINOR, REQ_VER_PATCH); err != nil {
		panic(fmt.Sprintf("Required PLC library version %d.%d.%d is not available", REQ_VER_MAJOR, REQ_VER_MINOR, REQ_VER_PATCH))
	}
}
