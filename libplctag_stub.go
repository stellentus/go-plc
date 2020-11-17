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
