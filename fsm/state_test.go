package fsm

import (
	"sync"
	"testing"
)

func TestMkID(t *testing.T) {
	// test with race detector
	const numRounds = 1000
	resetGID()
	var wg sync.WaitGroup
	wg.Add(numRounds)
	for i := 0; i < numRounds; i++ {
		go func() {
			mkID()
			wg.Done()
		}()
	}
	wg.Wait()

	id := mkID()
	if id != numRounds+1 {
		t.Fatalf("mkID: expected %d, got %d", numRounds+1, id)
	}
}
