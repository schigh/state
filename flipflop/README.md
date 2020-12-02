# flipflop

It's a fookin toggle state callback thinger.  There are no docs, but here's a usage example:

```go
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/schigh/state/flipflop"
)

const (
	wait = time.Second
	toggleState = uint(1)
	ocState = uint(2)
	ocState2 = uint(3)
	ocState3 = uint(4)
)

type service struct {
	name string
	ff   flipflop.FlipFlop
	down bool
	ts   int64
}

func dump(format string, i ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", i...)
}

func (s *service) handleToggle(ctx context.Context, _ bool) {

	now := time.Now().UnixNano()
	diff := time.Duration(now - s.ts)

	dump("%v - TOGGLE", diff)

	// <-time.After(wait)

	s.ts = time.Now().UnixNano()
	s.ff.Toggle(ctx, toggleState)
}

func (s *service) handleOC(ctx context.Context, closed bool) {
	now := time.Now().UnixNano()
	diff := time.Duration(now - s.ts)

	dump("%v - OC\tclosed: %t", diff, closed)

	// <-time.After(wait)

	if closed {
		s.ts = time.Now().UnixNano()
		s.ff.Open(ctx, ocState, ocState2, ocState3)
		return
	}
	s.ts = time.Now().UnixNano()
	s.ff.Close(ctx, ocState, ocState2, ocState3)
}

func main() {
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGTERM, syscall.SIGINT)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := &service{name: "s"}
	s.ff = flipflop.NewMachine(
		flipflop.WithSingleStateChangeHandler(s.handleToggle, toggleState),
		flipflop.WithSingleStateChangeHandler(s.handleOC, ocState),
		flipflop.WithSingleStateChangeHandler(s.handleOC, ocState2),
		flipflop.WithSingleStateChangeHandler(s.handleOC, ocState3),
	)
	s.ff.Run(ctx)

	s.ts = time.Now().UnixNano()
	s.ff.Toggle(ctx, toggleState)
	s.ff.Close(ctx, ocState, ocState2, ocState3)

	<-stopChan
}

```
