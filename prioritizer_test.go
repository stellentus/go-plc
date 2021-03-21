package plc

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestQuiescentEmptyPrioritizer(t *testing.T) {
	p := NewPrioritizer(1 * time.Microsecond)

	// Start() should kick off the worker goroutine and begin
	// boosting (of which there is nothing to boost, but this
	// ensures that we at least are not panicing)
	p.Start()
	time.Sleep(100 * time.Microsecond)
	p.Stop()
}

// This test ensures that we are able to execute a high priority
// task that has been scheduled before the Prioritizer starts.
func TestPrio0PrioritizerEnqueueBeforeStart(t *testing.T) {
	var i *uint64 = new(uint64)
	f := func() {
		atomic.AddUint64(i, 1)
	}

	p := NewPrioritizer(1 * time.Microsecond)

	p.Enqueue(f, 0)
	p.Start()

	time.Sleep(100 * time.Microsecond)
	assert.Greater(t, atomic.LoadUint64(i), uint64(0))
	p.Stop()
}

// This test ensures that we are able to execute a high priority
// task that has been scheduled before the Prioritizer starts.
func TestPrio0PrioritizerEnqueueAfterStart(t *testing.T) {
	var i *uint64 = new(uint64)
	f := func() {
		atomic.AddUint64(i, 1)
	}

	p := NewPrioritizer(1 * time.Microsecond)

	p.Start()

	time.Sleep(10 * time.Microsecond)
	p.Enqueue(f, 0)
	time.Sleep(90 * time.Microsecond)

	assert.Greater(t, atomic.LoadUint64(i), uint64(0))
	p.Stop()
}

// This test ensures that we are able to execute a high priority
// task that has been scheduled before the Prioritizer starts.
func TestNonZeroPriosPrioritizerEnqueueAfterStart(t *testing.T) {
	var i0 *uint64 = new(uint64)
	var i1 *uint64 = new(uint64)

	p := NewPrioritizer(10 * time.Microsecond)

	// We should see approximately 1/15th as many i1 increments as i0
	p.Enqueue(func() { atomic.AddUint64(i0, 1) }, 0)
	p.Enqueue(func() { atomic.AddUint64(i1, 1) }, 2)

	p.Start()
	time.Sleep(200 * time.Millisecond)

	assert.Greater(t, atomic.LoadUint64(i0), uint64(0))
	assert.Greater(t, atomic.LoadUint64(i1), uint64(0))
	assert.Greater(t, atomic.LoadUint64(i0), atomic.LoadUint64(i1))

	// Note: tracking the ratio of the low-prio task being scheduled to the high
	// becomes difficult in the presence of the adaptive boosting scheme.
	// It converges to 14.0 (TODO: why is it not 15.0, the ratio of prio0 to
	// prio2 tasks? The ratio always seems to be one less than the ratio I
	// expect, for instance, the ratio of prio-2 to prio-3 increments converge
	// to 3, not 4.) It seems to be doing a right thing, but not the right thing
	// I expected.
	r := float64(atomic.LoadUint64(i0)) / float64(atomic.LoadUint64(i1))
	assert.Greater(t, r, 13.9)
	// Seems to converge to 14.0
	assert.Less(t, r, 14.1)

	p.Stop()
}

// In this test, we schedule a bunch of high priority tasks that, in aggregate,
// exceed the per-second throughput for
func TestSlowPrio0StarvesLowerPrioTasks(t *testing.T) {
	var i0 *uint64 = new(uint64)
	var i1 *uint64 = new(uint64)

	p := NewPrioritizer(10 * time.Microsecond)

	// Servicing the prio-0 tasks requires at minimum 10 * 100ms, so we should never
	// have the extra time to boost a lower-priority task.
	for i := 0; i < 10; i++ {
		p.Enqueue(func() { time.Sleep(100 * time.Millisecond); atomic.AddUint64(i0, 1) }, 0)
	}
	p.Enqueue(func() { atomic.AddUint64(i1, 1) }, 1)

	p.Start()
	//time.Sleep(60 * time.Second)
	time.Sleep(2 * time.Second)

	p.mtx.Lock()
	assert.Equal(t, p.totalBoosts, uint64(0))
	p.mtx.Unlock()

	assert.Greater(t, atomic.LoadUint64(i0), uint64(0))
	assert.Equal(t, atomic.LoadUint64(i1), uint64(0))

	p.Stop()

}
