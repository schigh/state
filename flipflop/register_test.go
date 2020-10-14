package flipflop

import (
	"reflect"
	"testing"
)

func Test_offset(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		idx  uint
		wantIdx  uint
		wantOffs uint
		wantPanic bool
	}{
		{name: "0"},
		{name: "1", idx: 1, wantOffs: 1},
		{name: "2", idx: 2, wantOffs: 2},
		{name: "63", idx: 63, wantOffs: 63},
		{name: "64", idx: 64, wantIdx: 1, wantOffs: 0},
		{name: "65", idx: 65, wantIdx: 1, wantOffs: 1},
		{name: "3000", idx: 3000, wantIdx: 46, wantOffs: 56},
		{name: "4095", idx: 4095, wantIdx: 63, wantOffs: 63},
		{name: "4096", idx: 4096, wantIdx: 64, wantOffs: 0},
		{name: "4097", idx: 4097, wantPanic: true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			defer func(t *testing.T, p bool) {
				r := recover()
				if r == nil && p {
					t.Fatal("panic expected")
				}
				if r != nil && !p {
					t.Fatal("unexpected panic")
				}
			}(t, tt.wantPanic)
			got, got1 := offset(tt.idx)
			if got != tt.wantIdx {
				t.Errorf("offset() got = %v, want %v", got, tt.wantIdx)
			}
			if got1 != tt.wantOffs {
				t.Errorf("offset() got1 = %v, want %v", got1, tt.wantOffs)
			}
		})
	}
}

func Test_registerCloseAndAllClosed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		indices []uint
		chkIndices []uint
		want bool
	}{
		{
			name: "none",
			want: true,
		},
		{
			name: "1,2,3,4,5",
			indices: []uint{1,2,3,4,5},
			chkIndices: []uint{1,2,3,4,5},
			want: true,
		},
		{
			name: "1,2,3,4,5/6",
			indices: []uint{1,2,3,4,5},
			chkIndices: []uint{1,2,3,4,5,6},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var r register
			r, _ = registerClose(r, tt.indices...)
			if got := registerAllClosed(r, tt.chkIndices...); got != tt.want {
				t.Errorf("registerAllClosed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_registerOpenAllOpened(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		indices []uint
		chkIndices []uint
		want bool
	}{
		{
			name: "none",
			want: true,
		},
		{
			name: "1,2,3,4,5",
			indices: []uint{1,2,3,4,5},
			chkIndices: []uint{1,2,3,4,5},
			want: true,
		},
		{
			name: "1,2,3,4,5/6",
			indices: []uint{1,2,3,4,5},
			chkIndices: []uint{1,2,3,4,5,6},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := registerWithAllClosed()
			r, _ = registerOpen(r, tt.indices...)
			if got := registerAllOpened(r, tt.chkIndices...); got != tt.want {
				t.Errorf("registerAllOpened() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_registerAnyClosed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		indices []uint
		chkIndices []uint
		want bool
	}{
		{
			name: "none",
		},
		{
			name: "1,2,3,4,5",
			indices: []uint{1,2,3,4,5},
			chkIndices: []uint{5,6,7,8,9},
			want: true,
		},
		{
			name: "1,2,3,4,5-2",
			indices: []uint{1,2,3,4,5},
			chkIndices: []uint{6,7,8,9},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var r register
			r, _ = registerClose(r, tt.indices...)
			if got := registerAnyClosed(r, tt.chkIndices...); got != tt.want {
				t.Errorf("registerAnyClosed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_registerAnyOpened(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		indices []uint
		chkIndices []uint
		want bool
	}{
		{
			name: "none",
		},
		{
			name: "1,2,3,4,5",
			indices: []uint{1,2,3,4,5},
			chkIndices: []uint{5,6,7,8,9},
			want: true,
		},
		{
			name: "1,2,3,4,5-2",
			indices: []uint{1,2,3,4,5},
			chkIndices: []uint{6,7,8,9},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := registerWithAllClosed()
			r, _ = registerOpen(r, tt.indices...)
			if got := registerAnyOpened(r, tt.chkIndices...); got != tt.want {
				t.Errorf("registerAnyOpened() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_registerToggle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		preClose []uint
		preOpen []uint
		indices []uint
		wantClosed []uint
		wantOpened []uint
		start register
		want register
	}{
		{
			name: "none",
		},
		{
			name: "one closed",
			indices: []uint{1},
			wantClosed: []uint{1},
			want: func() register {
				var r register
				r, _ = registerClose(r, 1)
				return r
			}(),
		},
		{
			name: "several closed",
			indices: []uint{1,2,3,4,5,6,7,8,9},
			wantClosed: []uint{1,2,3,4,5,6,7,8,9},
			want: func() register {
				var r register
				r, _ = registerClose(r, 1,2,3,4,5,6,7,8,9)
				return r
			}(),
		},
		{
			name: "one opened",
			indices: []uint{1},
			wantOpened: []uint{1},
			start: registerWithAllClosed(),
			want: func() register {
				r, _ := registerOpen(registerWithAllClosed(), 1)
				return r
			}(),
		},
		{
			name: "several opened",
			indices: []uint{1,2,3,4,5,6,7,8,9},
			wantOpened: []uint{1,2,3,4,5,6,7,8,9},
			start: registerWithAllClosed(),
			want: func() register {
				r, _ := registerOpen(registerWithAllClosed(), 1,2,3,4,5,6,7,8,9)
				return r
			}(),
		},
		{
			name: "mixed bag 1",
			indices: []uint{1,2,3,4,5,6,7,8,9},
			wantClosed: []uint{1,3,5,7,9},
			wantOpened: []uint{2,4,6,8},
			start: func() register{
				var r register
				r, _ = registerClose(r, 2, 4, 6, 8)
				return r
			}(),
			want: func() register {
				var r register
				r, _ = registerClose(r, 1,3,5,7,9)
				return r
			}(),
		},
		{
			name: "mixed bag 2",
			indices: []uint{1,2,3,4,5,6,7,8,9},
			wantOpened: []uint{1,3,5,7,9},
			wantClosed: []uint{2,4,6,8},
			start: func() register{
				r, _ := registerOpen(registerWithAllClosed(), 2, 4, 6, 8)
				return r
			}(),
			want: func() register {
				r, _ := registerOpen(registerWithAllClosed(), 1,3,5,7,9)
				return r
			}(),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r, _ := registerClose(tt.start, tt.preClose...)
			r, _ = registerOpen(r, tt.preOpen...)
			got, gotClosed, gotOpened := registerToggle(r, tt.indices...)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("registerToggle() register = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(gotClosed, tt.wantClosed) {
				t.Errorf("registerToggle() closed indices = %v, want %v", gotClosed, tt.wantClosed)
			}
			if !reflect.DeepEqual(gotOpened, tt.wantOpened) {
				t.Errorf("registerToggle() opened indices = %v, want %v", gotOpened, tt.wantOpened)
			}
		})
	}
}
