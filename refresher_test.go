package plc

import (
	"testing"
	"time"
)

func TestRefresher(t *testing.T) {
	p := NewRefresher(newBlockingFake(), time.Millisecond)

	for range [5]int{} {
		go func() {
			i := 5
			p.ReadTag("bob", &i)
		}()
	}

}
