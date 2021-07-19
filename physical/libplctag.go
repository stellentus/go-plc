// +build !stub

package physical

/*
#cgo LDFLAGS: -lplctag
#include <libplctag.h>
*/
import "C"
import "fmt"

const (
	REQ_VER_MAJOR = 2
	REQ_VER_MINOR = 1
	REQ_VER_PATCH = 0
)

func init() {
	// Ensure the linked library uses the correct version
	err := newLibplctagError(C.plc_tag_check_lib_version(
		C.int(REQ_VER_MAJOR),
		C.int(REQ_VER_MINOR),
		C.int(REQ_VER_PATCH)))

	if err != nil {
		panic(fmt.Sprintf("Required PLC library version %d.%d.%d is not available",
			REQ_VER_MAJOR, REQ_VER_MINOR, REQ_VER_PATCH))
	}
}

const StubActive = false
