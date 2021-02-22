package plc

import (
	"sync"
)

type asyncer interface {
	Wait() error
	Add(name string, value interface{})
	AddError(err error)
}

type action func(string, interface{}) error

type job struct {
	name  string
	value interface{}
}

type async struct {
	action
	jobs    chan job
	cancel  chan struct{}
	wg      sync.WaitGroup
	errOnce sync.Once
	err     error
}

func newAsync(act action) *async {
	as := &async{
		action: act,
		jobs:   make(chan job, 1),
		cancel: make(chan struct{}),
	}

	as.wg.Add(1)
	go func() {
		// Listen for new jobs and launch them in new goroutines until the limit has been reached.
		numRoutines := 0
		var maxRoutines = 128

		defer as.wg.Done()
		for {
			select {
			case <-as.cancel:
				return
			case newJob, ok := <-as.jobs:
				if !ok {
					return
				}
				numRoutines++
				as.launchRoutine(newJob)
				if numRoutines >= maxRoutines {
					return
				}
			}
		}
	}()

	return as
}

func (as *async) launchRoutine(newJob job) {
	as.wg.Add(1)
	go func() {
		defer as.wg.Done()

		pendingJobs := true
		for pendingJobs {
			if err := as.takeAction(newJob); err != nil {
				return
			}
			newJob, pendingJobs = <-as.jobs
		}
	}()
}

func (as *async) takeAction(j job) error {
	if err := as.action(j.name, j.value); err != nil {
		as.setErr(err)
		return err
	}
	return nil
}

func (as *async) Wait() error {
	close(as.jobs) // No more should be sent

	done := make(chan struct{})
	go func() {
		as.wg.Wait()
		close(done)
	}()

	select {
	case <-as.cancel: // leaves some routines finishing up in the background
	case <-done: // completed successfully
	}

	return as.err
}

func (as *async) setErr(err error) {
	as.errOnce.Do(func() {
		as.err = err
		close(as.cancel)
		for _ = range as.jobs {
			// Drain jobs as quickly as possible since there was an error
		}
	})
}

func (as *async) Add(friendly string, value interface{}) {
	as.jobs <- job{friendly, value}
}

func (as *async) AddError(err error) {
	if err == nil {
		return
	}

	as.wg.Add(1)
	go func() {
		defer as.wg.Done()
		as.setErr(err)
	}()
}

type notAsync struct {
	error
	action
}

func newNotAsync(act action) *notAsync {
	return &notAsync{action: act}
}

func (nas *notAsync) Wait() error {
	return nas.error
}

func (nas *notAsync) Add(name string, value interface{}) {
	// Act immediately unless there's a cached error.
	if nas.error == nil {
		nas.error = nas.action(name, value)
	}
}
func (nas *notAsync) AddError(err error) {
	if nas.error == nil {
		nas.error = err
	}
}

type newAsyncer func(action) asyncer

func getNewAsyncer(useAsync bool) newAsyncer {
	if useAsync {
		return func(act action) asyncer { return newAsync(act) }
	} else {
		return func(act action) asyncer { return newNotAsync(act) }
	}
}
