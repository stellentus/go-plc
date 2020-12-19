package main

import (
	"flag"
	"io"
	"os"

	"github.com/stellentus/go-plc/l5x"
)

var (
	in  = flag.String("in", "l5x/test.L5X", "Path to L5X file to parse")
	out = flag.String("out", "", "Path to file to write generated types, or \"\" for stdout")
	pkg = flag.String("package", "", "If set, this string will be used as the package name. Otherwise, no package name will be printed.")
)

func main() {
	flag.Parse()
	var err error

	content, err := l5x.NewFromFile(*in)
	panicIfError(err, "Could not parse L5X '"+*in+"'")

	tl, err := content.Controller.TypeList()
	panicIfError(err, "Coundn't register ControlLogixTypes")

	fout := io.WriteCloser(os.Stdout)
	if *out != "" {
		fout, err = os.Create(*out)
		panicIfError(err, "Could not open output file '"+*out+"'")
	}
	defer fout.Close()

	// Print header line if requested
	if *pkg != "" {
		fout.Write([]byte("package " + *pkg + "\n\n"))
	}

	err = tl.WriteDefinitions(fout)
	panicIfError(err, "Failed to write definitions")
}

func panicIfError(err error, reason string) {
	if err != nil {
		panic("ERROR " + err.Error() + ": " + reason)
	}
}
