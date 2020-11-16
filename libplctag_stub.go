// +build stub

package plc

/*
#cgo LDFLAGS: ./libplctag_stub.a
#include "./libplctag.h"

LIB_EXPORT int plc_tag_lock(int32_t tag)
{
	return PLCTAG_STATUS_OK;
}
LIB_EXPORT int plc_tag_unlock(int32_t tag)
{
	return PLCTAG_STATUS_OK;
}

*/
import "C"

// CheckRequiredVersion returns an error if the version doesn't match the requirements.
func CheckRequiredVersion(major, minor, patch int) error {
	return newError(C.plc_tag_check_lib_version(C.int(major), C.int(minor), C.int(patch)))
}
