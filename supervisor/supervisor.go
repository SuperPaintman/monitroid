package supervisor

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/SuperPaintman/monitroid/gatherers"
)

type Result struct {
	Generation uint32      `json:"generation"`
	Success    interface{} `json:"success"`
	Error      error       `json:"error"`
}

type Supervisor struct {
	done chan struct{}

	mu        sync.RWMutex
	gatherers map[string]Result
}

func (s *Supervisor) Register(name string, d time.Duration, gatherer gatherers.Gatherer) {
	s.init()

	if _, ok := s.gatherers[name]; ok {
		panic(fmt.Errorf("supervisor: gatherer with name '%s' already registered", name))
	}

	s.gatherers[name] = Result{}

	go func() {
		timer := time.NewTimer(0)

		for {
			select {
			case <-s.done:
				return
			case <-timer.C:
			}

			res, err := gatherer.Gather()

			s.mu.Lock()
			gen := s.gatherers[name].Generation

			if err != nil {
				s.gatherers[name] = Result{
					Generation: gen + 1,
					Error:      err,
				}
			} else {
				s.gatherers[name] = Result{
					Generation: gen + 1,
					Success:    res,
				}
			}
			s.mu.Unlock()

			timer.Reset(d)
		}
	}()
}

func (s *Supervisor) DumpJSON(w io.Writer) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

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
