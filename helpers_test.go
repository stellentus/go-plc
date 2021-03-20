package plc

// test_helpers holds some mock datatypes and other useful contants that multiple test files
// may find helpful.

import (
	"sync"
	"time"
)

// Some shared constants acrossed reader/writer datatype tests.

const alphaCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const numCharset = "0123456789"
const alphanumCharset = alphaCharset + "_" + numCharset

const serialTestConcurrency = 1
const lowTestConcurrency = 10
const mediumTestConcurrency = 20
const highTestConcurrency = 30

const testWritePrecentage = 1
const heavyWritePerc = 50

// Stub ReadWriters

// latencyIntroducer simply delays processing a read or write request by some
// time interval. This is useful for artifically bounding throughput in
// concurrent tests.
type latencyIntroducer struct {
	downstream ReadWriter
	delay      time.Duration
}

func newLatencyIntroducer(downstream ReadWriter, delay time.Duration) *latencyIntroducer {
	return &latencyIntroducer{downstream, delay}
}

func (li *latencyIntroducer) ReadTag(name string, value interface{}) error {
	time.Sleep(li.delay)
	return li.downstream.ReadTag(name, value)
}

func (li *latencyIntroducer) WriteTag(name string, value interface{}) error {
	time.Sleep(li.delay)
	return li.downstream.WriteTag(name, value)
}

var _ = ReadWriter(&latencyIntroducer{})

// mockReadWriter is a dummy ReadWriter interface that unconditionally succeeds.
// We protect concurrenct accesses to the `state` map but _not_ the memory that
// the map's value points to, so it has to be the responsibility of the caller
// of ReadTag and WriteTag to ensure mutual exclusion on _their_ accesses.
// Typically, this will happen through a TagLocker wrapping this.
type mockReadWriter struct {
	state map[string]*uint32 // State is read/written to for the benefit of the race detector during unit tests.
	mtx   sync.Mutex
}

var _ = ReadWriter(&TagLocker{})

func newMockReadWriter() *mockReadWriter {
	return &mockReadWriter{
		state: make(map[string]*uint32),
	}
}
func (norw *mockReadWriter) ReadTag(name string, value interface{}) error {
	//fmt.Fprintf(os.Stderr, "Downstream: Reading %v\n", name)

	norw.mtx.Lock()
	ptr, ok := norw.state[name]
	if !ok {
		ptr = new(uint32)
		norw.state[name] = ptr
	}
	norw.mtx.Unlock()

	value = *ptr

	return nil
}
func (norw *mockReadWriter) WriteTag(name string, value interface{}) error {
	//fmt.Fprintf(os.Stderr, "Downstream: Writing %v\n", name)

	norw.mtx.Lock()
	ptr, ok := norw.state[name]
	if !ok {
		ptr = new(uint32)
		norw.state[name] = ptr
	}
	norw.mtx.Unlock()

	*ptr++

	return nil
}
