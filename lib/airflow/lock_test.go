package airflow

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Requests for the same DAG must never overlap. The previous implementation
// deleted the map entry on unlock, so a caller arriving while another still
// held the lock got a brand new mutex and ran concurrently.
func TestCallDagRequestLockSerializesSameDAG(t *testing.T) {
	const goroutines = 64

	var inCritical atomic.Int32
	var maxSeen atomic.Int32
	var wg sync.WaitGroup

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			unlock := callDagRequestLock("same-dag")
			defer unlock()

			n := inCritical.Add(1)
			for {
				max := maxSeen.Load()
				if n <= max || maxSeen.CompareAndSwap(max, n) {
					break
				}
			}
			// Widen the window so any overlap actually shows up.
			runtime.Gosched()
			inCritical.Add(-1)
		}()
	}
	wg.Wait()

	if got := maxSeen.Load(); got != 1 {
		t.Errorf("max concurrent holders for one DAG = %d, want 1", got)
	}
	assertNoLeakedEntries(t)
}

// Locks are per DAG, so unrelated DAGs must not serialize behind each other.
func TestCallDagRequestLockIsPerDAG(t *testing.T) {
	held := callDagRequestLock("dag-a")

	done := make(chan struct{})
	go func() {
		unlock := callDagRequestLock("dag-b")
		unlock()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("dag-b blocked on dag-a's lock; locks are not per DAG")
	}

	held()
	assertNoLeakedEntries(t)
}

// The map must not grow without bound: entries are dropped once the last
// caller releases them.
func TestCallDagRequestLockReleasesEntries(t *testing.T) {
	for i := 0; i < 10; i++ {
		unlock := callDagRequestLock("transient-dag")
		unlock()
	}
	assertNoLeakedEntries(t)
}

// Releasing twice must be a no-op rather than unlocking a mutex this caller no
// longer owns (which would panic or free a lock another caller holds).
func TestCallDagRequestLockUnlockIsIdempotent(t *testing.T) {
	unlock := callDagRequestLock("double-unlock-dag")
	unlock()
	unlock()

	assertNoLeakedEntries(t)

	// The lock must still be usable afterwards.
	again := callDagRequestLock("double-unlock-dag")
	again()
	assertNoLeakedEntries(t)
}

// Concurrent traffic across many DAGs exercises the shared map; run with -race.
func TestCallDagRequestLockConcurrentMapAccess(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		for _, dagID := range []string{"a", "b", "c"} {
			wg.Add(1)
			go func(id string) {
				defer wg.Done()
				unlock := callDagRequestLock(id)
				runtime.Gosched()
				unlock()
			}(dagID)
		}
	}
	wg.Wait()
	assertNoLeakedEntries(t)
}

func assertNoLeakedEntries(t *testing.T) {
	t.Helper()
	dagRequestsLock.Lock()
	n := len(dagRequests)
	dagRequestsLock.Unlock()
	if n != 0 {
		t.Errorf("dagRequests holds %d entries after every caller finished; want 0", n)
	}
}
