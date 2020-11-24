package plc

type ReadWriter interface {
	Reader
	Writer
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
	ReadWriter

	// Close closes the device.
	// The behavior of Close after the first call is undefined.
	// Specific implementations may document their own behavior.
	Close() error

	// GetList gets a list of tag names for the provided program
	// name (or all tags if no program name is provided).
	GetList(listName, prefix string) ([]Tag, []string, error)
}
