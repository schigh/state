package toggle

import (
	"context"
	"math/rand"
	"reflect"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestDelegate(t *testing.T) {
	t.Parallel()

	var (
		flip = func(t *testing.T, r register, adj uint, close bool) register {
			t.Helper()

			n := adj / CAP
			if close {
				r[n] |= 1 << (adj - (CAP * n))
				return r
			}

			r[n] &^= 1 << (adj - (CAP * n))
			return r
		}
	)

	t.Run("close", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name     string
			cancel   func()
			indices  []uint
			ctx      context.Context
			expected register
		}{
			{
				name: "close none",
				ctx:  context.Background(),
			},
			{
				name:    "close one",
				ctx:     context.Background(),
				indices: []uint{42},
				expected: func() register {
					var r register
					return flip(t, r, 42, true)
				}(),
			},
			{
				name:    "close some",
				ctx:     context.Background(),
				indices: []uint{5, 10, 15, 20},
				expected: func() register {
					var r register
					r = flip(t, r, 5, true)
					r = flip(t, r, 10, true)
					r = flip(t, r, 15, true)
					r = flip(t, r, 20, true)

					return r
				}(),
			},
			{
				name:    "close some more",
				ctx:     context.Background(),
				indices: []uint{2, 1, 2, 3, 2, 3, 7, 1, 4, 2, 9, 8, 2, 2, 3, 1},
				expected: func() register {
					var r register
					r = flip(t, r, 1, true)
					r = flip(t, r, 2, true)
					r = flip(t, r, 3, true)
					r = flip(t, r, 4, true)
					r = flip(t, r, 7, true)
					r = flip(t, r, 8, true)
					r = flip(t, r, 9, true)

					return r
				}(),
			},
		}

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				ctx, cancel := context.WithCancel(tt.ctx)
				defer cancel()

				d := newDelegate()
				d.close(ctx, tt.indices...)

				if !reflect.DeepEqual(d.reg, tt.expected) {
					expected := newDelegate()
					expected.reg = tt.expected
					t.Fatalf("delegate.close:\nexpected\n%s\n\ngot\n%s\n", expected.stringVal(), d.stringVal())
				}
			})
		}
	})

	t.Run("open", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name     string
			cancel   func()
			indices  []uint
			ctx      context.Context
			expected register
		}{
			{
				name:     "open none",
				ctx:      context.Background(),
				expected: registerWithAllClosed(),
			},
			{
				name:    "open one",
				ctx:     context.Background(),
				indices: []uint{42},
				expected: func() register {
					return flip(t, registerWithAllClosed(), 42, false)
				}(),
			},
			{
				name:    "open some",
				ctx:     context.Background(),
				indices: []uint{5, 10, 15, 20},
				expected: func() register {
					r := registerWithAllClosed()
					r = flip(t, r, 5, false)
					r = flip(t, r, 10, false)
					r = flip(t, r, 15, false)
					r = flip(t, r, 20, false)

					return r
				}(),
			},
			{
				name:    "open some more",
				ctx:     context.Background(),
				indices: []uint{2, 1, 2, 3, 2, 3, 7, 1, 4, 2, 9, 8, 2, 2, 3, 1},
				expected: func() register {
					r := registerWithAllClosed()
					r = flip(t, r, 1, false)
					r = flip(t, r, 2, false)
					r = flip(t, r, 3, false)
					r = flip(t, r, 4, false)
					r = flip(t, r, 7, false)
					r = flip(t, r, 8, false)
					r = flip(t, r, 9, false)

					return r
				}(),
			},
		}

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				ctx, cancel := context.WithCancel(tt.ctx)
				defer cancel()

				d := newDelegate()
				d.reg = registerWithAllClosed()
				d.open(ctx, tt.indices...)

				if !reflect.DeepEqual(d.reg, tt.expected) {
					expected := newDelegate()
					expected.reg = tt.expected
					t.Fatalf("delegate.open:\nexpected\n%s\n\ngot\n%s\n", expected.stringVal(), d.stringVal())
				}
			})
		}
	})

	// run with race detection
	t.Run("concurrency", func(t *testing.T) {
		t.Parallel()
		rand.Seed(time.Now().UnixNano())

		var pool sync.Pool
		pool.New = func() interface{} {
			return newDelegate()
		}

		var wg sync.WaitGroup

		// setting this number above 50 will cause a crash because
		// of the number of live goroutines
		// the supplied use case is more than ample to gauge load
		const numDelegates = 10

		mkIdx := func() uint {
			return uint(rand.Intn(4096))
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		for i := 0; i < numDelegates; i++ {
			wg.Add(1)
			go func(){
				var closeIt bool
				d := pool.Get().(*delegate)
				defer func(d *delegate) {
					d.reset()
					pool.Put(d)
				}(d)
				for j := 0; j < 100; j++ {
					if closeIt {
						d.close(ctx, mkIdx(), mkIdx(), mkIdx())
					} else {
						d.open(ctx, mkIdx(), mkIdx(), mkIdx())
					}

					closeIt = !closeIt
				}

				wg.Done()
			}()
		}

		wg.Wait()
	})

	t.Run("toggle", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())

		r := register{}
		r, _ = registerClose(r, 2,3,4,5,11,12,13,14,19,20)
		d := newDelegate()
		d.reg = r

		var opened, closed atomic.Value
		opened.Store([]uint{})
		closed.Store([]uint{})
		go func(ctx context.Context, d *delegate) {
			for {
				select {
				case <-ctx.Done():
					return
				case c := <-d.changeChan:
					if c.closed {
						cv := closed.Load().([]uint)
						cv = append(cv, c.state)
						closed.Store(cv)
						continue
					}
					ov := opened.Load().([]uint)
					ov = append(ov, c.state)
					opened.Store(ov)
				}
			}
		}(ctx, d)

		d.toggle(ctx, 1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20)

		<-time.After(100 * time.Millisecond)
		cancel()

		cv := closed.Load().([]uint)
		ov := opened.Load().([]uint)

		// they dont necessarily fire in order
		sort.SliceStable(cv, func(i, j int) bool {
			return cv[i] < cv[j]
		})
		sort.SliceStable(ov, func(i, j int) bool {
			return ov[i] < ov[j]
		})

		expectClosed := []uint{1,6,7,8,9,10,15,16,17,18}
		expectOpened := []uint{2,3,4,5,11,12,13,14,19,20}

		if !reflect.DeepEqual(expectClosed, cv) {
			t.Fatalf("toggle: expected closed:\n%v\ngot:\n%v\n", expectClosed, cv)
		}
		if !reflect.DeepEqual(expectOpened, ov) {
			t.Fatalf("toggle: expected opened:\n%v\ngot:\n%v\n", expectOpened, ov)
		}
	})
}

func TestStringVal(t *testing.T) {
	t.Parallel()

	const s1 = `63   0000000000000000000000000000000000000000000000000000000000000000
62   0000000000000000000000000000000000000000000000000000000000000000
61   0000000000000000000000000000000000000000000000000000000000000000
60   0000000000000000000000000000000000000000000000000000000000000000
59   0000000000000000000000000000000000000000000000000000000000000000
58   0000000000000000000000000000000000000000000000000000000000000000
57   0000000000000000000000000000000000000000000000000000000000000000
56   0000000000000000000000000000000000000000000000000000000000000000
55   0000000000000000000000000000000000000000000000000000000000000000
54   0000000000000000000000000000000000000000000000000000000000000000
53   0000000000000000000000000000000000000000000000000000000000000000
52   0000000000000000000000000000000000000000000000000000000000000000
51   0000000000000000000000000000000000000000000000000000000000000000
50   0000000000000000000000000000000000000000000000000000000000000000
49   0000000000000000000000000000000000000000000000000000000000000000
48   0000000000000000000000000000000000000000000000000000000000000000
47   0000000000000000000000000000000000000000000000000000000000000000
46   0000000000000000000000000000000000000000000000000000000000000000
45   0000000000000000000000000000000000000000000000000000000000000000
44   0000000000000000000000000000000000000000000000000000000000000000
43   0000000000000000000000000000000000000000000000000000000000000000
42   0000000000000000000000000000000000000000000000000000000000000000
41   0000000000000000000000000000000000000000000000000000000000000000
40   0000000000000000000000000000000000000000000000000000000000000000
39   0000000000000000000000000000000000000000000000000000000000000000
38   0000000000000000000000000000000000000000000000000000000000000000
37   0000000000000000000000000000000000000000000000000000000000000000
36   0000000000000000000000000000000000000000000000000000000000000000
35   0000000000000000000000000000000000000000000000000000000000000000
34   0000000000000000000000000000000000000000000000000000000000000000
33   0000000000000000000000000000000000000000000000000000000000000000
32   0000000000000000000000000000000000000000000000000000000000000000
31   0000000000000000000000000000000000000000000000000000000000000000
30   0000000000000000000000000000000000000000000000000000000000000000
29   0000000000000000000000000000000000000000000000000000000000000000
28   0000000000000000000000000000000000000000000000000000000000000000
27   0000000000000000000000000000000000000000000000000000000000000000
26   0000000000000000000000000000000000000000000000000000000000000000
25   0000000000000000000000000000000000000000000000000000000000000000
24   0000000000000000000000000000000000000000000000000000000000000000
23   0000000000000000000000000000000000000000000000000000000000000000
22   0000000000000000000000000000000000000000000000000000000000000000
21   0000000000000000000000000000000000000000000000000000000000000000
20   0000000000000000000000000000000000000000000000000000000000000000
19   0000000000000000000000000000000000000000000000000000000000000000
18   0000000000000000000000000000000000000000000000000000000000000000
17   0000000000000000000000000000000000000000000000000000000000000000
16   0000000000000000000000000000000000000000000000000000000000000000
15   0000000000000000000000000000000000000000000000000000000000000000
14   0000000000000000000000000000000000000000000000000000000000000000
13   0000000000000000000000000000000000000000000000000000000000000000
12   0000000000000000000000000000000000000000000000000000000000000000
11   0000000000000000000000000000000000000000000000000000000000000000
10   0000000000000000000000000000000000000000000000000000000000000000
9    0000000000000000000000000000000000000000000000000000000000000000
8    0000000000000000000000000000000000000000000000000000000000000000
7    0000000000000000000000000000000000000000000000000000000000000000
6    0000000000000000000000000000000000000000000000000000000000000000
5    0000000000000000000000000000000000000000000000000000000000000000
4    0000000000000000000000000000000000000000000000000000000000000000
3    0000000000000000000000000000000000000000000000000000000000000000
2    0000000000000000000000000000000000000000000000000000000000000000
1    0000000000000000000000000000000000000000000000000000000000000000
0    0000000000000000000000000000000000000000000000000000000000001111
`

	const s2 = `63   1111111111111111111111111111111111111111111111111111111111111111
62   1111111111111111111111111111111111111111111111111111111111111111
61   1111111111111111111111111111111111111111111111111111111111111111
60   1111111111111111111111111111111111111111111111111111111111111111
59   1111111111111111111111111111111111111111111111111111111111111111
58   1111111111111111111111111111111111111111111111111111111111111111
57   1111111111111111111111111111111111111111111111111111111111111111
56   1111111111111111111111111111111111111111111111111111111111111111
55   1111111111111111111111111111111111111111111111111111111111111111
54   1111111111111111111111111111111111111111111111111111111111111111
53   1111111111111111111111111111111111111111111111111111111111111111
52   1111111111111111111111111111111111111111111111111111111111111111
51   1111111111111111111111111111111111111111111111111111111111111111
50   1111111111111111111111111111111111111111111111111111111111111111
49   1111111111111111111111111111111111111111111111111111111111111111
48   1111111111111111111111111111111111111111111111111111111111111111
47   1111111111111111111111111111111111111111111111111111111111111111
46   1111111111111111111111111111111111111111111111111111111111111111
45   1111111111111111111111111111111111111111111111111111111111111111
44   1111111111111111111111111111111111111111111111111111111111111111
43   1111111111111111111111111111111111111111111111111111111111111111
42   1111111111111111111111111111111111111111111111111111111111111111
41   1111111111111111111111111111111111111111111111111111111111111111
40   1111111111111111111111111111111111111111111111111111111111111111
39   1111111111111111111111111111111111111111111111111111111111111111
38   1111111111111111111111111111111111111111111111111111111111111111
37   1111111111111111111111111111111111111111111111111111111111111111
36   1111111111111111111111111111111111111111111111111111111111111111
35   1111111111111111111111111111111111111111111111111111111111111111
34   1111111111111111111111111111111111111111111111111111111111111111
33   1111111111111111111111111111111111111111111111111111111111111111
32   1111111111111111111111111111111111111111111111111111111111111111
31   1111111111111111111111111111111111111111111111111111111111111111
30   1111111111111111111111111111111111111111111111111111111111111111
29   1111111111111111111111111111111111111111111111111111111111111111
28   1111111111111111111111111111111111111111111111111111111111111111
27   1111111111111111111111111111111111111111111111111111111111111111
26   1111111111111111111111111111111111111111111111111111111111111111
25   1111111111111111111111111111111111111111111111111111111111111111
24   1111111111111111111111111111111111111111111111111111111111111111
23   1111111111111111111111111111111111111111111111111111111111111111
22   1111111111111111111111111111111111111111111111111111111111111111
21   1111111111111111111111111111111111111111111111111111111111111111
20   1111111111111111111111111111111111111111111111111111111111111111
19   1111111111111111111111111111111111111111111111111111111111111111
18   1111111111111111111111111111111111111111111111111111111111111111
17   1111111111111111111111111111111111111111111111111111111111111111
16   1111111111111111111111111111111111111111111111111111111111111111
15   1111111111111111111111111111111111111111111111111111111111111111
14   1111111111111111111111111111111111111111111111111111111111111111
13   1111111111111111111111111111111111111111111111111111111111111111
12   1111111111111111111111111111111111111111111111111111111111111111
11   1111111111111111111111111111111111111111111111111111111111111111
10   1111111111111111111111111111111111111111111111111111111111111111
9    1111111111111111111111111111111111111111111111111111111111111111
8    1111111111111111111111111111111111111111111111111111111111111111
7    1111111111111111111111111111111111111111111111111111111111111111
6    1111111111111111111111111111111111111111111111111111111111111111
5    1111111111111111111111111111111111111111111111111111111111111111
4    1111111111111111111111111111111111111111111111111111111111111111
3    1111111111111111111111111111111111111111111111111111111111111111
2    1111111111111111111111111111111111111111111111111111111111111111
1    1111111111111111111111111111111111111111111111111111111111111111
0    1111111111111111111111111111111111111111111111111111111111110000
`

	d1 := newDelegate()
	d2 := newDelegate()
	d2.reg = registerWithAllClosed()

	d1.close(context.Background(), 0,1,2,3)
	d2.open(context.Background(), 0,1,2,3)

	if d1.stringVal() != s1 {
		t.Fatalf("stringval: expected:\n%s\n\ngot:\n%s\n\n", s1, d1.stringVal())
	}

	if d2.stringVal() != s2 {
		t.Fatalf("stringval: expected:\n%s\n\ngot:\n%s\n\n", s2, d2.stringVal())
	}
}

func TestContextCancellation(t *testing.T) {
	t.Parallel()

	t.Run("close", func(t *testing.T) {
		t.Parallel()

		t.Run("cancel", func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			d := newDelegate()
			d.close(ctx, 0, 1, 2)
			cancel()
			d.close(ctx, 3, 4, 5)

			r := register{}
			r, _ = registerClose(r, 0, 1, 2)

			if !reflect.DeepEqual(r, d.reg) {
				t.Fatalf("register mismatch:\nexpected:\n%#v\ngot:\n%#v\n", r, d.reg)
			}
		})
		t.Run("deadline exceeded", func(t *testing.T) {
			t.Parallel()

			ctx, _ := context.WithTimeout(context.Background(), 100 * time.Millisecond)
			d := newDelegate()
			d.close(ctx, 0, 1, 2)
			<-time.After(110 * time.Millisecond)
			d.close(ctx, 3, 4, 5)

			r := register{}
			r, _ = registerClose(r, 0, 1, 2)

			if !reflect.DeepEqual(r, d.reg) {
				t.Fatalf("register mismatch:\nexpected:\n%#v\ngot:\n%#v\n", r, d.reg)
			}
		})
	})
	t.Run("open", func(t *testing.T) {
		t.Parallel()

		t.Run("cancel", func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			d := newDelegate()
			d.reg = registerWithAllClosed()
			d.open(ctx, 0, 1, 2)
			cancel()
			d.open(ctx, 3, 4, 5)

			r := registerWithAllClosed()
			r, _ = registerOpen(r, 0, 1, 2)

			if !reflect.DeepEqual(r, d.reg) {
				t.Fatalf("register mismatch:\nexpected:\n%#v\ngot:\n%#v\n", r, d.reg)
			}
		})
		t.Run("deadline exceeded", func(t *testing.T) {
			t.Parallel()

			ctx, _ := context.WithTimeout(context.Background(), 100 * time.Millisecond)
			d := newDelegate()
			d.reg = registerWithAllClosed()
			d.open(ctx, 0, 1, 2)
			<-time.After(110 * time.Millisecond)
			d.open(ctx, 3, 4, 5)

			r := registerWithAllClosed()
			r, _ = registerOpen(r, 0, 1, 2)

			if !reflect.DeepEqual(r, d.reg) {
				t.Fatalf("register mismatch:\nexpected:\n%#v\ngot:\n%#v\n", r, d.reg)
			}
		})
	})
	t.Run("toggle", func(t *testing.T) {
		t.Parallel()

		t.Run("cancel", func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			d := newDelegate()
			d.toggle(ctx, 0,1,2,3,4,5,6,7)
			d.toggle(ctx, 5,6,7,8,9)
			cancel()
			d.toggle(ctx, 10,11,12,13)

			r := register{}
			r, _ = registerClose(r, 0,1,2,3,4,8,9)

			if !reflect.DeepEqual(r, d.reg) {
				t.Fatalf("register mismatch:\nexpected:\n%#v\ngot:\n%#v\n", r, d.reg)
			}
		})
		t.Run("deadline exceeded", func(t *testing.T) {
			t.Parallel()

			ctx, _ := context.WithTimeout(context.Background(), 100 * time.Millisecond)
			d := newDelegate()
			d.toggle(ctx, 0,1,2,3,4,5,6,7)
			d.toggle(ctx, 5,6,7,8,9)
			<-time.After(110 * time.Millisecond)
			d.toggle(ctx, 10,11,12,13)

			r := register{}
			r, _ = registerClose(r, 0,1,2,3,4,8,9)

			if !reflect.DeepEqual(r, d.reg) {
				t.Fatalf("register mismatch:\nexpected:\n%#v\ngot:\n%#v\n", r, d.reg)
			}
		})
	})
}
