package plc

type ReadWriter interface {
	Reader
	Writer
}

// Reader writes values from a PLC.
type Reader interface {
	// ReadTag reads the requested tag into the provided value.
	ReadTag(name string, value interface{}) error

	// ReadTagAtIndex reads the requested array tag at the given index into the provided value.
	// It's provided to be faster than ReadTag when only a single array element is needed
	ReadTagAtIndex(name string, index int, value interface{}) error
}

// Writer writes values out to a PLC.
type Writer interface {
	// WriteTag writes the provided tag and value.
	WriteTag(name string, value interface{}) error

	// WriteTagAtIndex writes the requested array tag at the given index with the provided value.
	// It's provided to be faster than WriteTag when only a single array element is needed. (Otherwise
	// would be necessary to read into an entire slice, edit one element, and re-write the slice,
	// which is not atomic.
	WriteTagAtIndex(name string, index int, value interface{}) error
}
