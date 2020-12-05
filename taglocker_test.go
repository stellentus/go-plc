package plc

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"testing"
)

const alphaCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const numCharset = "0123456789"
const alphanumCharset = alphaCharset + "_" + numCharset

// mockReadWriter is a dummy ReadWriter interface that unconditionally succeeds.  We
// protect concurrenct accesses to the `state` map but _not_ the memory that the map's
// value points to, so it has to be the responsibility of the caller of ReadTag and
// WriteTag to ensure mutual exclusion on _their_ accesses.
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

// Generates a random fully-qualified tag, and returns
// paths to that tag but also all intermediary prefixes.
// We will test that we can lock intermediary nodes by
// randomly choosing from these returned values.
func genPaths() []string {
	var paths []string

	/*
	* The EBNF is:
	*
	* tag ::= SYMBOLIC_SEG ( tag_seg )* ( bit_seg )?
	*
	* tag_seg ::= '.' SYMBOLIC_SEG
	*             '[' array_seg ']'
	*
	* bit_seg ::= '.' [0-9]+
	*
	* array_seg ::= NUMERIC_SEG ( ',' NUMERIC_SEG )*
	*
	* SYMBOLIC_SEG ::= [a-zA-Z]([a-zA-Z0-9_]*)
	*
	* NUMERIC_SEG ::= [0-9]+
	*
	 */
	genSymbolicSeg := func() string {
		length := 1 + rand.Intn(10)
		bs := make([]byte, length)
		bs[0] = alphaCharset[rand.Intn(len(alphaCharset))]
		for i := 1; i < len(bs); i++ {
			bs[i] = alphanumCharset[rand.Intn(len(alphanumCharset))]
		}
		return string(bs)
	}

	genNumericSeg := func() string {
		return fmt.Sprintf("%d", rand.Intn(10))
	}
	genArraySeg := func() string {
		// TODO: when the parser supports comma-separated multidimensional
		// arrays, include potentially generating those.
		return genNumericSeg()
	}
	genTagSeg := func() string {
		if rand.Intn(2) == 0 {
			return "." + genSymbolicSeg()
		}
		return fmt.Sprintf("[%s]", genArraySeg())
	}
	// tag ::= SYMBOLIC_SEG ( tag_seg )* ( bit_seg )?
	var sb strings.Builder
	sb.WriteString(genSymbolicSeg())
	for i := 0; i < rand.Intn(5); i++ {
		//paths = append(paths, sb.String())
		sb.WriteString(genTagSeg())
	}
	paths = append(paths, sb.String())
	return paths
}

const serialConcurrency = 1
const lowConcurrency = 10
const medConcurrency = 20
const highConcurrency = 30
const writePerc = 1
const heavyWritePerc = 50

const numTags = 10

func BenchmarkSerialTagLocking(b *testing.B) {
	benchmarkTagLockLocking(b, serialConcurrency, writePerc)
}

func BenchmarkLowConcurrencyTagLocking(b *testing.B) {
	benchmarkTagLockLocking(b, lowConcurrency, writePerc)
}

func BenchmarkMedConcurrencyTagLocking(b *testing.B) {
	benchmarkTagLockLocking(b, medConcurrency, writePerc)
}

func BenchmarkHighConcurrencyTagLocking(b *testing.B) {
	benchmarkTagLockLocking(b, highConcurrency, writePerc)
}

func BenchmarkSerialTagLockingWithHeavyWrites(b *testing.B) {
	benchmarkTagLockLocking(b, serialConcurrency, heavyWritePerc)
}

func BenchmarkLowConcurrencyTagLockingWithHeavyWrites(b *testing.B) {
	benchmarkTagLockLocking(b, lowConcurrency, heavyWritePerc)
}

func BenchmarkMedConcurrencyTagLockingWithHeavyWrites(b *testing.B) {
	benchmarkTagLockLocking(b, medConcurrency, heavyWritePerc)
}
func BenchmarkHighConcurrencyTagLockingWithHeavyWrites(b *testing.B) {
	benchmarkTagLockLocking(b, highConcurrency, heavyWritePerc)
}

func benchmarkTagLockLocking(b *testing.B, concurrency int, writePerc int) {
	barrier := make(chan bool, concurrency)

	tl := NewTagLocker(newMockReadWriter())

	read := func(name string) {
		var garbage uint32

		err := tl.ReadTag(name, garbage)
		if err != nil {
			b.Errorf("%v\n", err)
		}
		<-barrier
	}
	write := func(name string) {
		err := tl.WriteTag(name, 42)
		if err != nil {
			b.Errorf("%v\n", err)
		}
		<-barrier
	}

	// Generate `numTags` randomly-named tags; the slice of tags
	// stores those tags but also all intermediary prefixes to
	// those tags, so larger structures can be locked.
	var tags []string
	for i := 0; i < numTags; i++ {
		tags = append(tags, genPaths()...)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tag := tags[rand.Intn(len(tags))]
		isWrite := rand.Intn(100) < writePerc

		barrier <- true
		if isWrite {
			go write(tag)
		} else {
			go read(tag)
		}
	}
	for {
		select {
		case <-barrier:
		default:
			b.StopTimer()
			return
		}
	}
}
