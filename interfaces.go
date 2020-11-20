package plc

type ReadWriter interface {
	Reader
	Writer
}

// Reader writes values from a PLC.
type Reader interface {
	// ReadTag reads the requested tag into the provided value.
	ReadTag(name string, value interface{}) error
}

// Writer writes values out to a PLC.
type Writer interface {
	// WriteTag writes the provided tag and value.
	WriteTag(name string, value interface{}) error
}

// rawDevice is an interface to a PLC device.
type rawDevice interface {
	// ReadTag reads the requested tag into the provided value.
	ReadTag(name string, value interface{}) error

	// WriteTag writes the provided tag and value.
	WriteTag(name string, value interface{}) error

	// Close cleans up resources.
	Close() error

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
