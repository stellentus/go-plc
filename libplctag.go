// +build !stub

package plc

/*
#cgo LDFLAGS: ./libplctag.a
#include "./libplctag.h"
*/
import "C"

// CheckRequiredVersion returns an error if the version doesn't match the requirements.
func CheckRequiredVersion(major, minor, patch int) error {
	return newError(C.plc_tag_check_lib_version(C.int(major), C.int(minor), C.int(patch)))
}
