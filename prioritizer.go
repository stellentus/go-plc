package plc

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	movingaverage "github.com/RobinUS2/golang-moving-average"
)

/* 0-6 */
const priorities = 7

var priorityFactors [6]uint64 = [6]uint64{
	1,                 /* 1 -> 0 */
	15,                /* 2 -> 1 */
	60,                /* 3 -> 2 */
	60 * 60,           /* 4 -> 3 */
	60 * 60 * 60,      /* 5 -> 4 */
	60 * 60 * 60 * 24, /* 6 -> 5 */
}

// runnable encapsulates the state necessary to repeatedly schedule
// a task at some priority level.
type runnable struct {
	// The operation to run.
	f task

	// The non-boosted priority for this task.  After f() executes,
	// the Prioritizer will re-enqeue this task at this level again.
	initialPrio int
}

// a runqueue is a queue of tasks to run, and an associated factor that
// indicates with what frequency tasks should be boosted to the next
// higher priority level.
type runqueue []runnable

// Prioritizer handles periodic scheduling of tasks according to priority levels:
//
// Priority 0: = sample / 1 sec
// Priority 1: ~ sample / 1 sec
// Priority 2: ~ sample / 15 sec
// Priority 3: ~ sample / 60 sec
// Priority 4: ~ sample / (60 * 60) sec
// Priority 5: ~ sample / (60 * 60 * 60) sec
// Priority 6: ~ sample / (60 * 60 * 60 * 24) sec
//
// By default, priority-0 tags are inserted into the ready-to-run queue.
// Lower priority tags are stored in a per-prior
//
// Our tasks will be, but the implementation does not assume to be, functions
// that poll a Reader for the value of some tag.
//
// The Prioritizer will periodically execute tasks on the the ready-to-run
// queue, and continuously estimate its throughput. Whatever buffer time remains
// after servicing all the Priority 0 tasks is used to additionally sample
// lower-priority tasks.
//
// The delta between the number of priority zero tasks serviced and the number of
// tasks that can be serviced at all is enqueued from the next lower priority
// level. For layers below that, a proprtional priority boost also takes place
// (i.e. for every 15 excutions, a task from priority 2 is boosted to priority 1,
// and for every four executions, a task from priority 3 is boosted to priority 2,
// and so on.)
type Prioritizer struct {
	// Ensures mutual exclusion on Prioritizer fields, since a user
	// can call Enqueue() concurrently with runOne() running in the worker
	// goroutine.
	mtx sync.Mutex

	// All non-zero tasks to execute, indexed by their current priority level.
	// The tasks that are ready to run are enqueued at runqueue[0].
	runqueues [priorities]runqueue

	// How many total requests have been serviced?
	totalBoosts uint64

	// How many priority-0 tasks have been enqueued by the user?  We use this to
	// estimate "mandatory throughput".  As a result, it is okay if they are updated
	// non-atomically with tasks (but still require the lock to be held to avoid
	// spurious races).
	prioZeroTasks uint
	totalTasks    uint

	// A maximal pause to wait in between servicing requests on the run queue.
	// Should be a small nonzero value to ensure we aren't completely hammering
	// downstream readers and thrashing the Go scheduler.
	sleepinterval time.Duration

	// Should we keep executing the main loop?
	running bool
}

//
func (p *Prioritizer) boost() {
	p.totalBoosts++
	// Based on the boost factor between each priority level, potentially
	// promote a lower-priority task up a priority.
	for i := len(priorityFactors) - 1; i >= 0; i-- {
		if p.totalBoosts%priorityFactors[i] != 0 {
			continue
		}

		r, ok := p.popBack(i + 1)
		if ok {
			p.pushFront(r, i)
		}
	}
}

// pushFront enqueues the supplied runnable into the appropriate runqueue.
func (p *Prioritizer) pushFront(r runnable, prio int) {
	p.mtx.Lock()
	p.runqueues[prio] = append(p.runqueues[prio], r)
	p.mtx.Unlock()
}

// pushFront enqueues the supplied runnable into the appropriate runqueue.
func (p *Prioritizer) popBack(prio int) (runnable, bool) {
	var r runnable

	p.mtx.Lock()
	defer func() {
		p.mtx.Unlock()
	}()

	if len(p.runqueues[prio]) == 0 {
		return r, false
	}

	r = p.runqueues[prio][0]
	p.runqueues[prio] = p.runqueues[prio][1:]

	return r, true
}

// runOne attempts to execute a task on the ready-to-run queue.  If no elements
// are enqueued, it will continuously boost lower-priority tasks until one has
// been enqueued.
func (p *Prioritizer) runOne() {
	// Is the runqueue empty for whatever reason? If so, boost a lower-priority
	// task and try again.  (Typically speaking this should only ever happen as
	// the system is initialising and the constructing goroutine is still
	// inserting runnables, or if there are no priority-0 tasks.)
	var r runnable
	var ok bool
	for !ok {
		r, ok = p.popBack(0)
		if !ok {
			p.boost()
			runtime.Gosched()
		}
	}

	// r is a valid runnable - run it and re-enqueue it at its _original_ priority
	// level.
	r.f()
	p.pushFront(r, r.initialPrio)
}

// Start kicks off the dispatcher goroutine.
func (p *Prioritizer) Start() {
	go func() {
		// A running average of how long each job takes to complete, in microseconds.
		// TODO: how big of an average should we have?
		avg := movingaverage.New(128)

		running := true

		for running {
			b := time.Now()
			p.runOne()
			e := time.Now()

			// Recompute the running average.  If our tasks/sec exceeds the number
			// of prio-0 tasks, we can boo
			avg.Add(float64(e.Sub(b).Microseconds()))

			p.mtx.Lock()
			p0Tasks := p.prioZeroTasks
			nonP0tasks := p.totalTasks - p.prioZeroTasks
			p.mtx.Unlock()

			if p.prioZeroTasks > 0 {
				// The boost factor is the ratio of the nubmer of tasks we can
				// complete within a second to the number of tasks we must.
				// (Don't bother boosting more than the total number of non-p0
				// tasks.)
				// XXX: Note that if we are consistently not able to keep up with
				// the p0 tasks, we will effectively be starving lower-priority
				// tasks.  We should consider an occasional "mercy-boost" so
				// things get through at least asymptotically-slowly...
				boostFactor := uint(1/avg.Avg()/1e-6) / p0Tasks
				if boostFactor > nonP0tasks {
					boostFactor = nonP0tasks
				}
				var i uint
				for ; i < boostFactor; i++ {
					p.boost()
				}
			}

			// How much time do we have left in our time quantum?  Sleep that much.
			remaining := time.Now().Add(p.sleepinterval).Sub(b)
			if remaining.Microseconds() > 0 {
				time.Sleep(remaining)
			}

			// Exit if we've been instructed to stop running.
			p.mtx.Lock()
			running = p.running
			p.mtx.Unlock()
		}
	}()
}

// Stop ceases the Prioritizer's scheduler loop.
func (p *Prioritizer) Stop() {
	p.mtx.Lock()
	p.running = false
	p.mtx.Unlock()
}

// Enqueue will schedule a particular task at the supplied priority level.
func (p *Prioritizer) Enqueue(t task, prio int) error {
	if prio < 0 || uint64(prio) >= priorities {
		return fmt.Errorf("Invalid priority %d: must be on [0, %d)", prio, priorities)
	}

	p.mtx.Lock()
	if prio == 0 {
		p.prioZeroTasks++
	}
	p.totalTasks++
	p.mtx.Unlock()

	r := runnable{
		f:           t,
		initialPrio: prio,
	}
	p.pushFront(r, prio)

	return nil
}

// NewPrioritizer produces a new Prioritizer.
func NewPrioritizer(interval time.Duration) *Prioritizer {
	p := &Prioritizer{
		sleepinterval: interval,
		running:       true,
	}
	for i := 0; i < len(p.runqueues); i++ {
		p.runqueues[i] = make(runqueue, 0)
	}

	return p
}
