package plc

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
)

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

func BenchmarkSerialTagLocking(b *testing.B) {
	benchmarkTagLockLocking(b, serialTestConcurrency, testWritePrecentage)
}

func BenchmarkLowConcurrencyTagLocking(b *testing.B) {
	benchmarkTagLockLocking(b, lowTestConcurrency, testWritePrecentage)
}

func BenchmarkMedConcurrencyTagLocking(b *testing.B) {
	benchmarkTagLockLocking(b, mediumTestConcurrency, testWritePrecentage)
}

func BenchmarkHighConcurrencyTagLocking(b *testing.B) {
	benchmarkTagLockLocking(b, highTestConcurrency, testWritePrecentage)
}

func BenchmarkSerialTagLockingWithHeavyWrites(b *testing.B) {
	benchmarkTagLockLocking(b, serialTestConcurrency, heavyWritePerc)
}

func BenchmarkLowConcurrencyTagLockingWithHeavyWrites(b *testing.B) {
	benchmarkTagLockLocking(b, lowTestConcurrency, heavyWritePerc)
}

func BenchmarkMedConcurrencyTagLockingWithHeavyWrites(b *testing.B) {
	benchmarkTagLockLocking(b, mediumTestConcurrency, heavyWritePerc)
}
func BenchmarkHighConcurrencyTagLockingWithHeavyWrites(b *testing.B) {
	benchmarkTagLockLocking(b, highTestConcurrency, heavyWritePerc)
}

func benchmarkTagLockLocking(b *testing.B, concurrency int, writePerc int) {
	const numTags = 10

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
