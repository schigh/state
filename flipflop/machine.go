package flipflop

import (
	"context"
	"fmt"
)

type ChangeHandler func(ctx context.Context, idx uint, state bool)
type SingleStateChangeHandler func(ctx context.Context, state bool)

type FlipFlop interface {
	fmt.GoStringer
	Open(ctx context.Context, conditions ...uint)
	Close(ctx context.Context, conditions ...uint)
	Toggle (ctx context.Context, conditions ...uint)
	Run(ctx context.Context)
}

type flipFlop struct {
	delegate      *delegate
	defaultChange func(context.Context, uint, bool)
	changeMap     map[uint]func(context.Context, bool)
}

var _ FlipFlop = (*flipFlop)(nil)

type Option func(*flipFlop)

func WithDefaultChangeHandler(handler ChangeHandler) Option {
	return func(m *flipFlop) {
		m.defaultChange = handler
	}
}

func WithSingleStateChangeHandler(handler SingleStateChangeHandler, condition uint) Option {
	return func(m *flipFlop) {
		m.changeMap[condition] = handler
	}
}

func WithAllStatesClosed() Option {
	return func(m *flipFlop) {
		defer m.delegate.lock().unlock()
		m.delegate.reg = registerWithAllClosed()
	}
}

func NewMachine(opts ...Option) *flipFlop {
	m := flipFlop{
		delegate:      newDelegate(),
		defaultChange: func(context.Context, uint, bool) {},
		changeMap:     make(map[uint]func(context.Context, bool)),
	}

	for _, f := range opts {
		f(&m)
	}

	return &m
}

func (ff *flipFlop) Run(ctx context.Context) {
	go func(ctx context.Context, m *flipFlop) {
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
	}(ctx, ff)
}

func (ff *flipFlop) Close(ctx context.Context, conditions ...uint) {
	ff.delegate.close(ctx, conditions...)
}

func (ff *flipFlop) Open(ctx context.Context, conditions ...uint) {
	ff.delegate.open(ctx, conditions...)
}

func (ff *flipFlop) Toggle(ctx context.Context, conditions ...uint) {
	ff.delegate.toggle(ctx, conditions...)
}

func (ff *flipFlop) GoString() string {
	return ff.delegate.stringVal()
}
