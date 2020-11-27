// +build stub

package plc

/*
#cgo LDFLAGS: -lplctagstub
#include <libplctag.h>
*/
import "C"

const dummyTagType = 0x2000

func init() {
	// The stub uses type 0x2000 for everything. The pre-loaded ones are uint16.
	RegisterTagTypeName(dummyTagType, "dummy_type")
	RegisterTagType(dummyTagType, uint16(0))
}
