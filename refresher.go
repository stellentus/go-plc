package plc

/*
#include <stdint.h>
*/
import "C"
import (
	"reflect"
	"sync"
	"time"
)

var tagTypes = map[C.uint16_t]interface{}{
	2: "",
}

// A Refresher can be used to periodically reissue the read for every seen value, so that values are readily available in a cache.
type Refresher struct {
	plc    Reader
	period time.Duration
	seen   map[string]struct{}
	mutex  sync.Mutex
}

var _ = Reader(&Refresher{}) // Compiler makes sure this type is a Reader

// NewRefresher returns a refresher that will update every read value.
func NewRefresher(plc Reader, period time.Duration) *Refresher {
	return &Refresher{
		plc:    plc,
		period: period,
		seen:   map[string]struct{}{},
	}
}

func (r *Refresher) launchIfNecessary(name string, value interface{}, f func(v interface{})) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if _, ok := r.seen[name]; ok {
		return
	}

	r.seen[name] = struct{}{}
	value = reflect.New(reflect.TypeOf(value))

	go func() {
		for _ = range time.NewTicker(r.period).C {
			f(value)
		}
	}()
}

func (r *Refresher) ReadTag(name string, value interface{}) error {
	r.launchIfNecessary(name, value, func(v interface{}) { r.plc.ReadTag(name, v) })

	return r.plc.ReadTag(name, value)
}
