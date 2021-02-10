package plc

// Pooled wraps another plc.ReadWriter with a work pool that runs a set number of concurrent operations.
type Pooled struct {
	plc         ReadWriter
	read, write tasker
}

var _ = ReadWriter(Pooled{}) // Compiler makes sure this type is a ReadWriter

// NewPooled creates a new Pooled and launches worker goroutines to handle incoming reads and writes.
// There is no way to kill the workers once they're launched.
func NewPooled(plc ReadWriter, workers int) Pooled {
	read, write := make(tasker), make(tasker)
	for i := 0; i < workers; i++ {
		go worker(read, write)
	}
	return Pooled{plc, read, write}
}

func (p Pooled) ReadTag(name string, value interface{}) error {
	return p.read.task(func() error { return p.plc.ReadTag(name, value) })
}

func (p Pooled) WriteTag(name string, value interface{}) error {
	return p.write.task(func() error { return p.plc.WriteTag(name, value) })
}

type task func()
type tasker chan task

func (t tasker) task(f func() error) error {
	ch := make(chan error)
	t <- func() { ch <- f() }
	return <-ch
}

func worker(read, write <-chan task) {
	for {
		select {
		case t := <-write:
			t()
		case t := <-read:
			t()
		}
	}
}
