package toggle

import (
	"context"
	"fmt"
)

type ChangeHandler func(ctx context.Context, idx uint, state bool)
type SingleStateChangeHandler func(ctx context.Context, state bool)

type FSM interface {
	fmt.GoStringer
	Open(ctx context.Context, conditions ...uint)
	Close(ctx context.Context, conditions ...uint)
	Toggle (ctx context.Context, conditions ...uint)
	Run(ctx context.Context)
}

type machine struct {
	delegate      *delegate
	defaultChange func(context.Context, uint, bool)
	changeMap     map[uint]func(context.Context, bool)
}

var _ = FSM(&machine{})

type Option func(*machine)

func WithDefaultChangeHandler(handler ChangeHandler) Option {
	return func(m *machine) {
		m.defaultChange = handler
	}
}

func WithSingleStateChangeHandler(handler SingleStateChangeHandler, condition uint) Option {
	return func(m *machine) {
		m.changeMap[condition] = handler
	}
}

func WithAllStatesClosed() Option {
	return func(m *machine) {
		defer m.delegate.lock().unlock()
		m.delegate.reg = registerWithAllClosed()
	}
}

func NewMachine(opts ...Option) *machine {
	m := machine{
		delegate:      newDelegate(),
		defaultChange: func(context.Context, uint, bool) {},
		changeMap:     make(map[uint]func(context.Context, bool)),
	}

	for _, f := range opts {
		f(&m)
	}

	return &m
}

func (m *machine) Run(ctx context.Context) {
	go func(ctx context.Context, m *machine) {
		for {
			select {
			case <-ctx.Done():
				return
			case c := <-m.delegate.changeChan:
				if f, ok := m.changeMap[c.state]; ok {
					go f(c.ctx, c.closed)
					continue
				}
				go m.defaultChange(c.ctx, c.state, c.closed)
			}
		}
	}(ctx, m)
}

func (m *machine) Close(ctx context.Context, conditions ...uint) {
	m.delegate.close(ctx, conditions...)
}

func (m *machine) Open(ctx context.Context, conditions ...uint) {
	m.delegate.open(ctx, conditions...)
}

func (m *machine) Toggle(ctx context.Context, conditions ...uint) {
	m.delegate.toggle(ctx, conditions...)
}

func (m *machine) GoString() string {
	return m.delegate.stringVal()
}
