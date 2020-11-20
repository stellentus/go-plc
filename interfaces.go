package plc

type ReadWriter interface {
	Reader
	Writer
}

type ReadWriteCloser interface {
	Reader
	Writer
	Closer
}

// Reader is the interface that wraps the basic ReadTag method.
type Reader interface {
	// ReadTag reads the requested tag into the provided value.
	ReadTag(name string, value interface{}) error
}

// Writer is the interface that wraps the basic WriteTag method.
type Writer interface {
	// WriteTag writes the provided tag and value.
	WriteTag(name string, value interface{}) error
}

// Closer is the interface that wraps the basic Close method.
//
// The behavior of Close after the first call is undefined.
// Specific implementations may document their own behavior.
type Closer interface {
	Close() error
}

// rawDevice is an interface to a PLC device.
type rawDevice interface {
	ReadWriteCloser

	// GetList gets a list of tag names for the provided program
	// name (or all tags if no program name is provided).
	GetList(listName, prefix string) ([]Tag, []string, error)
}

// ReaderFunc is a function that can be used as a Reader.
// It's the same pattern as http.HandlerFunc.
type ReaderFunc func(name string, value interface{}) error

func (f ReaderFunc) ReadTag(name string, value interface{}) error {
	return f(name, value)
}

// WriterFunc is a function that can be used as a Writer.
// It's the same pattern as http.HandlerFunc.
type WriterFunc func(name string, value interface{}) error

func (f WriterFunc) WriteTag(name string, value interface{}) error {
	return f(name, value)
}
