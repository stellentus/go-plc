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
