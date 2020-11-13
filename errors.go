package plc

/*
#include "./libplctag.h"
*/
import "C"

type errorCode struct {
	code C.int32_t
}

func (err errorCode) Error() string {
	cstr := C.plc_tag_decode_error(C.int(err.code))
	return C.GoString(cstr)
}

// Pending is an error indicating the PLC has not yet provided a result for the non-blocking request.
type Pending struct {
	errorCode
}

func newError(code C.int32_t) error {
	err := errorCode{code: code}
	if code == 0 {
		return nil
	} else if code > 0 {
		return Pending{err}
	} else {
		return err
	}
}

const (
	pendingStatus = C.PLCTAG_STATUS_PENDING
	okStatus      = C.PLCTAG_STATUS_OK

	abortError          = C.PLCTAG_ERR_ABORT
	badConfigError      = C.PLCTAG_ERR_BAD_CONFIG
	badConnectionError  = C.PLCTAG_ERR_BAD_CONFIG
	badDataError        = C.PLCTAG_ERR_BAD_DATA
	badDeviceError      = C.PLCTAG_ERR_BAD_DEVICE
	badGatewayError     = C.PLCTAG_ERR_BAD_GATEWAY
	badParamError       = C.PLCTAG_ERR_BAD_PARAM
	badReplyError       = C.PLCTAG_ERR_BAD_REPLY
	badStatusError      = C.PLCTAG_ERR_BAD_STATUS
	closeError          = C.PLCTAG_ERR_CLOSE
	createError         = C.PLCTAG_ERR_CREATE
	duplicateError      = C.PLCTAG_ERR_DUPLICATE
	encodeError         = C.PLCTAG_ERR_ENCODE
	mutexDestroyError   = C.PLCTAG_ERR_MUTEX_DESTROY
	mutexInitError      = C.PLCTAG_ERR_MUTEX_INIT
	mutexLockError      = C.PLCTAG_ERR_MUTEX_LOCK
	mutexUnlockError    = C.PLCTAG_ERR_MUTEX_UNLOCK
	notAllowedError     = C.PLCTAG_ERR_NOT_ALLOWED
	notFoundError       = C.PLCTAG_ERR_NOT_FOUND
	notImplementedError = C.PLCTAG_ERR_NOT_IMPLEMENTED
	noDataError         = C.PLCTAG_ERR_NO_DATA
	noMatchError        = C.PLCTAG_ERR_NO_MATCH
	noMemError          = C.PLCTAG_ERR_NO_MEM
	noResourcesError    = C.PLCTAG_ERR_NO_RESOURCES
	nullPtrError        = C.PLCTAG_ERR_NULL_PTR
	openError           = C.PLCTAG_ERR_OPEN
	outOfBoundsError    = C.PLCTAG_ERR_OUT_OF_BOUNDS
	readError           = C.PLCTAG_ERR_READ
	remoteErrError      = C.PLCTAG_ERR_REMOTE_ERR
	threadCreateError   = C.PLCTAG_ERR_THREAD_CREATE
	threadJoinError     = C.PLCTAG_ERR_THREAD_JOIN
	timeoutError        = C.PLCTAG_ERR_TIMEOUT
	tooLargeError       = C.PLCTAG_ERR_TOO_LARGE
	tooSmallError       = C.PLCTAG_ERR_TOO_SMALL
	unsupportedError    = C.PLCTAG_ERR_UNSUPPORTED
	winsockError        = C.PLCTAG_ERR_WINSOCK
	writeError          = C.PLCTAG_ERR_WRITE
	partialError        = C.PLCTAG_ERR_PARTIAL
)
