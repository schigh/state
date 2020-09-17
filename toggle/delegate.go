package toggle

import (
	"context"
	"fmt"
	"strings"
)

type delegate struct {
	locker     chan struct{}
	changeChan chan change
	reg        register
}

type change struct {
	ctx    context.Context
	state  uint
	closed bool
}

func newDelegate() *delegate {
	sem := make(chan struct{}, 1)
	sem <- struct{}{}
	return &delegate{
		locker:     sem,
		changeChan: make(chan change),
	}
}

func (d *delegate) lock() *delegate {
	<-d.locker
	return d
}

func (d *delegate) unlock() {
	d.locker <- struct{}{}
}

func (d *delegate) close(ctx context.Context, indices ...uint) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	defer d.lock().unlock()

	var changes []uint
	var r register
	r, changes = registerClose(d.reg, indices...)
	d.reg = r

	for i := 0; i < len(changes); i++ {
		go d.pushChange(ctx, changes[i], true)
	}
}

func (d *delegate) open(ctx context.Context, indices ...uint) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	defer d.lock().unlock()

	var changes []uint
	var r register
	r, changes = registerOpen(d.reg, indices...)
	d.reg = r

	for i := 0; i < len(changes); i++ {
		go d.pushChange(ctx, changes[i], false)
	}
}

func (d *delegate) toggle(ctx context.Context, indices ...uint) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	defer d.lock().unlock()

	var opened, closed []uint
	var r register
	r, closed, opened = registerToggle(d.reg, indices...)
	d.reg = r

	for i := 0; i < len(closed); i++ {
		go d.pushChange(ctx, closed[i], true)
	}
	for i := 0; i < len(opened); i++ {
		go d.pushChange(ctx, opened[i], false)
	}
}

func (d *delegate) pushChange(ctx context.Context, state uint, closed bool) {
	d.changeChan <- change{ctx, state, closed}
}

func (d *delegate) reset() {
	defer d.lock().unlock()
	d.reg = register{}
}

func (d *delegate) stringVal() string {
	defer d.lock().unlock()
	sb := strings.Builder{}

	for i := CAP - 1; i >= 0; i-- {
		sb.WriteString(fmt.Sprintf("%-5d%064b\n", i, d.reg[i]))
	}

	return sb.String()
}
