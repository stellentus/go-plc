package plc

type readerFunc func(string, interface{}) error

func (rd readerFunc) ReadTag(name string, value interface{}) error {
	return rd(name, value)
}

type writerFunc func(string, interface{}) error

func (wr writerFunc) WriteTag(name string, value interface{}) error {
	return wr(name, value)
}
