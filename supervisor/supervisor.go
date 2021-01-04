package supervisor

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/SuperPaintman/monitroid/gatherers"
)

var ErrNotReady = errors.New("supervisor: gatherer not ready")

type JSONError struct {
	Err error
}

func (e JSONError) MarshalJSON() ([]byte, error) {
	if e.Err == nil {
		return []byte("null"), nil
	}

	s := e.Err.Error()

	return json.Marshal(s)
}

type Result struct {
	Generation uint32      `json:"generation"`
	Success    interface{} `json:"success"`
	Error      JSONError   `json:"error"`
}

func (r Result) Err() error {
	return r.Error.Err
}

type Observer interface {
	Observe(name string, gatherer gatherers.Gatherer, result Result)
}

var _ Observer = (ObserverFunc)(nil)

type ObserverFunc func(name string, gatherer gatherers.Gatherer, result Result)

func (fn ObserverFunc) Observe(name string, gatherer gatherers.Gatherer, result Result) {
	fn(name, gatherer, result)
}

type Supervisor struct {
	done chan struct{}

	muObservers sync.RWMutex
	observers   []Observer

	muGatherers sync.RWMutex
	gatherers   map[string]Result
}

func (s *Supervisor) Register(name string, d time.Duration, gatherer gatherers.Gatherer) {
	s.init()

	if _, ok := s.gatherers[name]; ok {
		panic(fmt.Errorf("supervisor: gatherer with name '%s' already registered", name))
	}

	s.gatherers[name] = Result{
		Error: JSONError{
			Err: ErrNotReady,
		},
	}

	go func() {
		timer := time.NewTimer(0)

		for {
			select {
			case <-s.done:
				return
			case <-timer.C:
			}

			res, err := gatherer.Gather()

			s.muGatherers.RLock()
			gen := s.gatherers[name].Generation
			s.muGatherers.RUnlock()

			// TODO(SuperPaintman): check if result has changed.

			var result Result
			if err != nil {
				result = Result{
					Generation: gen + 1,
					Error: JSONError{
						Err: err,
					},
				}
			} else {
				result = Result{
					Generation: gen + 1,
					Success:    res,
				}
			}

			s.muGatherers.Lock()
			s.gatherers[name] = result
			s.muGatherers.Unlock()

			s.muObservers.RLock()
			for _, o := range s.observers {
				o.Observe(name, gatherer, result)
			}
			s.muObservers.RUnlock()

			timer.Reset(d)
		}
	}()
}

func (s *Supervisor) Observe(o Observer) {
	s.muObservers.Lock()
	defer s.muObservers.Unlock()

	s.observers = append(s.observers, o)
}

func (s *Supervisor) DumpJSON(w io.Writer) error {
	s.muGatherers.RLock()
	defer s.muGatherers.RUnlock()

	result := struct {
		Gatherers map[string]Result `json:"gatherers"`
	}{
		Gatherers: s.gatherers,
	}

	if err := json.NewEncoder(w).Encode(&result); err != nil {
		return fmt.Errorf("supervisor: failed to encode json: %w", err)
	}

	return nil
}

func (s *Supervisor) Stop() {
	if s.done != nil {
		close(s.done)
	}
}

func (s *Supervisor) init() {
	if s.done == nil {
		s.done = make(chan struct{})
	}

	if s.gatherers == nil {
		s.gatherers = make(map[string]Result)
	}
}
