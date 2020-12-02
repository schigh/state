package fsm

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/schigh/slice"
)

type machine struct {
	mu sync.RWMutex
	curr atomic.Value
	start atomic.Value
	endStates map[uint64]State
	idx uint32
	transitions map[uint64][]Transition
	cancel func()
}

type Option func(*machine)

func WithTransitions(transitions ...Transition) Option {
	return func(m *machine) {
		m.mu.Lock()
		defer m.mu.Unlock()

		if m.transitions == nil {
			m.transitions = make(map[uint64][]Transition)
		}

		// this sets the first state in the first transition as the root
		if len(transitions) > 0 {
			m.start.Store(transitions[0].From())
		}

		for _, t := range transitions {
			id := t.From().Id()
			m.transitions[id] = append(m.transitions[id], t)
		}
	}
}

func NewMachine(opts ...Option) *machine {
	m := machine{}
	for _, f := range opts {
		f(&m)
	}

	return &m
}

func (m *machine) Graph() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// TODO: traverse the fsm and generate graph
}

func (m *machine) SetStart(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var start State
	for _, t := range m.transitions {
		tt := TransitionSlice(t)
		tName := tt.Filter(func(tr Transition) bool {
			return tr.From().Name() == name
		})
		if len(tName) > 0 {
			start = tName[0].From()
			break
		}
	}

	if start == nil {
		return fmt.Errorf("no state found with name: %s", name)
	}

	m.start.Store(start)
	m.curr.Store(start)

	return nil
}

func (m *machine) Reset() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	starti := m.start.Load()
	if starti == nil {
		return errors.New("this machine has no start state")
	}
	start, _ := starti.(State)
	m.curr.Store(start)

	return nil
}

func (m *machine) SetEndStates(names  ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.endStates == nil {
		m.endStates = make(map[uint64]State)
	}

	validNames := make(map[string]bool)

	for _, t := range m.transitions {
		tt := TransitionSlice(t)
		tt.Filter(func(tr Transition) bool {
			return slice.String(names).Contains(tr.To().Name())
		}).Each(func(tr Transition) {
			m.endStates[tr.To().Id()] = tr.To()
			validNames[tr.To().Name()] = true
		})
	}

	idx, ok := slice.String(names).IfEach(func(s string) bool {
		_, ok := validNames[s]
		return ok
	})

	if !ok {
		return fmt.Errorf("invalid state: '%s'", names[idx])
	}


	return nil
}

func (m *machine) IsEndState() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.endStates == nil {
		m.endStates = make(map[uint64]State)
		return false
	}

	curr, _ := m.curr.Load().(State)
	_, ok := m.endStates[curr.Id()]

	return ok
}

func (m *machine) AddTransition(t Transition) {
	m.mu.Lock()
	defer m.mu.Unlock()

	from := t.From()
	if from == nil || t.To() == nil {
		panic("transition must have a FROM and TO state")
	}

	if m.transitions == nil {
		m.transitions = make(map[uint64][]Transition)
		m.start.Store(from)
	}

	id := from.Id()
	m.transitions[id] = append(m.transitions[id], t)
}

func (m *machine) Validate() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sm := make(map[uint64]State)

	start, _ := m.start.Load().(State)
	if start == nil {
		return errors.New("no start state set")
	}
	var stateNames []string
	for _, tt := range m.transitions {
		for _, t := range tt {
			if t.From() == nil {
				return fmt.Errorf("transition '%s' has no from state", t.Description())
			}
			if t.To() == nil {
				return fmt.Errorf("transition '%s' has no to state", t.Description())
			}
			sm[t.From().Id()] = t.From()
			sm[t.To().Id()] = t.To()
			stateNames = append(stateNames, t.From().Name(), t.To().Name())
		}
	}

	stateNames = slice.String(stateNames).Unique()

	if len(stateNames) != len(sm) {
		return errors.New("invalid: all state names must be unique")
	}

	return nil
}

func (m *machine) Current() State {
	m.mu.RLock()
	defer m.mu.RUnlock()

	curr, _ := m.curr.Load().(State)
	if curr == nil {
		curr, _ = m.start.Load().(State)
	}
	return curr
}

func (m *machine) Update(ctx context.Context, value interface{}) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}

	curr, _ := m.curr.Load().(State)
	if curr == nil {
		curr, _ = m.start.Load().(State)
		if curr == nil {
			return false, errors.New("machine has no start state")
		}
	}
	transitions, ok := m.transitions[curr.Id()]
	if !ok {
		return false, nil
	}

	for _, t := range transitions {
		success, err := t.Go(ctx, value)
		if err != nil {
			return false, err
		}

		if success {
			to := t.To()
			if to != nil {
				m.curr.Store(to)
			}

			return true, nil
		}
	}

	return false, nil
}
