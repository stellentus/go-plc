package main

import (
	"flag"
	"fmt"

	"github.com/stellentus/go-plc/l5x"
)

var (
	path = flag.String("path", "l5x/test.L5X", "Path to L5X file to parse")
)

func main() {
	content, err := l5x.NewFromFile(*path)
	panicIfError(err, "Could not open path '"+*path+"'")

	fmt.Println(*content)
}

func panicIfError(err error, reason string) {
	if err != nil {
		panic("ERROR " + err.Error() + ": " + reason)
	}
}
