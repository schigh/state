package fsm

import (
	"context"
	"sync/atomic"
)

//go:generate slicify Transition all

var gid uint64                    // nolint:gochecknoglobals
var lock = make(chan struct{}, 1) // nolint:gochecknoglobals

func resetGID() {
	atomic.StoreUint64(&gid, 0)
}

func mkID() uint64 {
	lock <- struct{}{}
	defer func() {
		<-lock
	}()
	id := atomic.AddUint64(&gid, 1)
	return id
}

type Trigger uint

type TriggerFunc func(context.Context, interface{}) (bool, error)

type Identifier interface {
	Id() uint64
}

type State interface {
	Identifier
	Name() string
	When(string, TriggerFunc) Transition
}

type Transition interface {
	Identifier
	Description() string
	From() State
	To() State
	Then(State) Transition
	Go(context.Context, interface{}) (bool, error)
}

type machineState struct {
	name string
	id   uint64
}

type StateOption func(machineState) machineState

func NewState(name string, options ...StateOption) State {
	s := machineState{id: mkID(), name: name}
	for _, f := range options {
		s = f(s)
	}

	return s
}

func (s machineState) Name() string {
	return s.name
}

func (s machineState) When(desc string, f TriggerFunc) Transition {
	return &edge{id: mkID(), from: s, f: f, desc: desc}
}

func (s machineState) Id() uint64 {
	return s.id
}

type edge struct {
	desc string
	from State
	to   State
	f    TriggerFunc
	id   uint64
}

func (e *edge) Id() uint64 {
	return e.id
}

func (e *edge) Description() string {
	return e.desc
}

func (e *edge) From() State {
	return e.from
}

func (e *edge) To() State {
	return e.to
}

func (e *edge) Then(s State) Transition {
	e.to = s
	return e
}

func (e *edge) Go(ctx context.Context, v interface{}) (bool, error) {
	return e.f(ctx, v)
}
