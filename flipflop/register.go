package flipflop

import (
	"fmt"
	"math"
)

const (
	CAP = 64
	WORDSIZE = 64
	MAXREG = CAP * WORDSIZE
)

func offset(idx uint) (uint, uint) {
	if idx > MAXREG {
		panic(fmt.Sprintf("state: overflow - a single state flipFlop can hold no more than %d indices", MAXREG))
	}

	return idx / WORDSIZE, idx % WORDSIZE
}

func shift(offs uint) uint64 {
	return uint64(1 << offs)
}

type register [CAP]uint64

func registerWithAllClosed() (r register) {
	for i := 0; i < CAP; i++ {
		r[i] = math.MaxUint64
	}
	return
}

// func registerUnion(left, right register) (out register) {
// 	for i := 0; i < CAP; i++ {
// 		out[i] = left[i] & right[i]
// 	}
// 	return
// }
//
// func registerDiff(left, right register) (out register) {
// 	for i := 0; i < CAP; i++ {
// 		out[i] = left[i] ^ right[i]
// 	}
// 	return
// }

func registerClose(r register, indices ...uint) (register, []uint) {
	out := make([]uint, 0, len(indices))
	for i := 0; i < len(indices); i++ {
		idx, offs := offset(indices[i])
		shifted := shift(offs)
		if r[idx]&shifted != shifted {
			out = append(out, indices[i])
		}
		r[idx] |= shifted
	}

	return r, out[:]
}

func registerClosed(r register, index uint) bool {
	idx, offs := offset(index)
	shifted := shift(offs)
	return r[idx]& shifted == shifted
}

func registerAllClosed(r register, indices ...uint) bool {
	for i := 0; i < len(indices); i++ {
		if !registerClosed(r, indices[i]) {
			return false
		}
	}

	return true
}

func registerAnyClosed(r register, indices ...uint) bool {
	for i := 0; i < len(indices); i++ {
		if registerClosed(r, indices[i]) {
			return true
		}
	}

	return false
}

func registerOpen(r register, indices ...uint) (register, []uint) {
	out := make([]uint, 0, len(indices))
	for i := 0; i < len(indices); i++ {
		idx, offs := offset(indices[i])
		shifted := shift(offs)
		if r[idx]&shifted == shifted {
			out = append(out, indices[i])
		}
		r[idx] &^= shifted
	}

	return r, out[:]
}

func registerOpened(r register, index uint) bool {
	idx, offs := offset(index)
	shifted := shift(offs)
	return r[idx]& shifted != shifted
}

func registerAllOpened(r register, indices ...uint) bool {
	for i := 0; i < len(indices); i++ {
		if !registerOpened(r, indices[i]) {
			return false
		}
	}

	return true
}

func registerAnyOpened(r register, indices ...uint) bool {
	for i := 0; i < len(indices); i++ {
		if registerOpened(r, indices[i]) {
			return true
		}
	}

	return false
}

func registerToggle(r register, indices ...uint) (register, []uint, []uint) {
	var opened, closed []uint
	for i := 0; i < len(indices); i++ {
		idx, offs := offset(indices[i])
		shifted := shift(offs)

		if r[idx]&shifted != shifted {
			closed = append(closed, indices[i])
			r[idx] |= shifted
		} else {
			opened = append(opened, indices[i])
			r[idx] &^= shifted
		}
	}

	return r, closed[:], opened[:]
}
