package plc

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func BenchmarkSerialPooledOperations(b *testing.B) {
	benchmarkPooledOperations(b, serialTestConcurrency, 0)
}

func BenchmarkLowConcurrentPooledOperations(b *testing.B) {
	benchmarkPooledOperations(b, lowTestConcurrency, 0)
}

func BenchmarkMediumConcurrentPooledOperations(b *testing.B) {
	benchmarkPooledOperations(b, mediumTestConcurrency, 0)
}

func BenchmarkHighConcurrentPooledOperations(b *testing.B) {
	benchmarkPooledOperations(b, highTestConcurrency, 0)
}

func benchmarkPooledOperations(b *testing.B, concurrency int, writePerc int) {
	const numTags = 10
	const delay = 1 * time.Millisecond

	barrier := make(chan bool, concurrency)

	// p exposes concurrent accesses to a mock read/writer, with an artificial
	// delay introduced between the two of them.
	// XXX: This leaks goroutines.  So it goes.
	p := NewPooled(newLatencyIntroducer(newMockReadWriter(), delay), concurrency)

	read := func(name string) {
		var garbage uint32

		err := p.ReadTag(name, garbage)
		if err != nil {
			b.Errorf("%v\n", err)
		}
		<-barrier
	}

	var tags []string
	for i := 0; i < numTags; i++ {
		tags = append(tags, fmt.Sprintf("DUMMY_AQUA_TAG_%d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tag := tags[rand.Intn(len(tags))]

		barrier <- true
		go read(tag)
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
