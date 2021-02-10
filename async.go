package plc

import (
	"golang.org/x/sync/errgroup"
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

var noJob = job{}

type async struct {
	action
	*errgroup.Group
	jobs chan job
}

func newAsync(act action) async {
	as := async{
		action: act,
		Group:  &errgroup.Group{},
		jobs:   make(chan job, 1),
	}

	as.launchRoutine(noJob)

	go func() {
		// Listen for new jobs and launch them in new goroutines until the limit has been reached.
		numRoutines := 1 // One was already started above
		var maxRoutines = 128
		for newJob := range as.jobs {
			numRoutines++
			as.launchRoutine(newJob)
			if numRoutines >= maxRoutines {
				return
			}
		}
	}()

	return as
}

func (as async) launchRoutine(newJob job) {
	as.Group.Go(func() error {
		if newJob != noJob {
			// First handle the provided job.
			if err := as.takeAction(newJob); err != nil {
				return err
			}
		}

		// Now keep listening for new jobs
		for j := range as.jobs {
			if err := as.takeAction(j); err != nil {
				return err
			}
		}

		return nil
	})
}

func (as async) takeAction(j job) error {
	if err := as.action(j.name, j.value); err != nil {
		for _ = range as.jobs {
			// Drain jobs as quickly as possible since there was an error
		}
		return err
	}
	return nil
}

func (as async) Wait() error {
	close(as.jobs)
	err := as.Group.Wait()
	return err
}

func (as async) Add(name string, value interface{}) {
	as.jobs <- job{name, value}
}

func (as async) AddError(err error) {
	as.Go(func() error {
		for _ = range as.jobs {
			// Drain jobs as quickly as possible since there was an error
		}
		return err
	})
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
