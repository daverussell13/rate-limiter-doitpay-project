package util_test

import (
	"sync"
	"testing"
	"time"

	"github.com/daverussell13/rate-limiter-doitpay-project/internal/util"
)

func TestStripedMutex_BasicLockUnlock(t *testing.T) {
	sm := util.NewStripedMutex(16)

	key := "test-key"
	unlock := sm.Lock(key)
	done := make(chan bool)
	go func() {
		sm.Lock(key)()
		done <- true
	}()

	select {
	case <-done:
		t.Fatal("lock should have blocked, but it didn't")
	case <-time.After(50 * time.Millisecond):
	}

	unlock()

	select {
	case <-done:
	case <-time.After(50 * time.Millisecond):
		t.Fatal("lock was not released properly")
	}
}

func TestStripedMutex_ParallelDifferentKeys(t *testing.T) {
	sm := util.NewStripedMutex(16)
	var wg sync.WaitGroup
	keys := []string{"a", "b", "c", "d", "e"}
	counter := 0
	mu := sync.Mutex{}

	for _, k := range keys {
		wg.Add(1)
		go func(key string) {
			defer wg.Done()
			unlock := sm.Lock(key)
			defer unlock()

			time.Sleep(10 * time.Millisecond)
			mu.Lock()
			counter++
			mu.Unlock()
		}(k)
	}

	wg.Wait()
	if counter != len(keys) {
		t.Fatalf("expected counter=%d, got %d", len(keys), counter)
	}
}
