package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// sync/atomic — lock-free atomic counter
// Faster than Mutex for simple numeric operations.
// No locks, no blocking — uses CPU-level atomic instructions.

func atomicCounter() {

	var counter atomic.Int64
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Add(1) // atomic increment, no mutex needed
		}()
	}

	wg.Wait()
	fmt.Printf("  Counter: %d (expected 1000)\n", counter.Load())
}

// atomic.Bool — thread-safe flag
// Great for stop signals, feature flags, ready-state checks.

func atomicBool() {

	var running atomic.Bool
	running.Store(true)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		iterations := 0
		for running.Load() {
			iterations++
			time.Sleep(10 * time.Millisecond)
		}
		fmt.Printf("  Worker did %d iterations before stop\n", iterations)
	}()

	time.Sleep(55 * time.Millisecond)
	running.Store(false) // signal stop without channels or mutexes
	wg.Wait()
}

// atomic.Value — store and load arbitrary values atomically
// Useful for config hot-reloading.

type Config struct {
	MaxConns int
	Timeout  time.Duration
}

func atomicValue() {

	var cfg atomic.Value
	cfg.Store(Config{MaxConns: 10, Timeout: 5 * time.Second})

	// Reader goroutines — always get a consistent snapshot
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			c := cfg.Load().(Config)
			fmt.Printf("  Reader %d sees: MaxConns=%d, Timeout=%v\n", id, c.MaxConns, c.Timeout)
		}(i)
	}

	// Writer updates the config (readers are never blocked)
	cfg.Store(Config{MaxConns: 50, Timeout: 10 * time.Second})
	fmt.Println("  Config updated!")

	wg.Wait()
}

// sync.Map — concurrent map from the standard library
// Best when keys are stable (read-heavy, rare writes) or
// when many goroutines read/write disjoint key sets.
// For most cases, a regular map + RWMutex is preferred.

func syncMap() {

	var m sync.Map
	var wg sync.WaitGroup

	// Writers — each goroutine writes its own key
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := fmt.Sprintf("worker-%d", id)
			m.Store(key, id*100)
		}(i)
	}
	wg.Wait()

	// Read all entries
	fmt.Println("  Entries:")
	m.Range(func(key, value any) bool {
		fmt.Printf("    %s = %v\n", key, value)
		return true // return false to stop iteration
	})

	// LoadOrStore — atomic "get or set" in one call
	actual, loaded := m.LoadOrStore("worker-0", 999)
	fmt.Printf("  LoadOrStore worker-0: val=%v, alreadyExisted=%v\n", actual, loaded)

	actual2, loaded2 := m.LoadOrStore("worker-new", 999)
	fmt.Printf("  LoadOrStore worker-new: val=%v, alreadyExisted=%v\n", actual2, loaded2)
}

// sync.Pool — reuse temporary objects to reduce GC pressure
// Objects may be collected by GC at any time, so don't rely on persistence.
// Great for buffers, encoders, decoders, etc.

func syncPool() {

	pool := &sync.Pool{
		New: func() any {
			fmt.Println("  [Pool] Creating new buffer")
			return make([]byte, 0, 1024)
		},
	}

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Get a buffer from pool (or create a new one)
			buf := pool.Get().([]byte)
			buf = buf[:0] // reset length

			// Use the buffer
			buf = append(buf, fmt.Sprintf("data-from-worker-%d", id)...)
			fmt.Printf("  Worker %d used buffer: %s\n", id, string(buf))

			// Return buffer to pool for reuse
			pool.Put(buf)
		}(i)
	}
	wg.Wait()
}

// sync.Cond — condition variable
// Lets goroutines wait for a condition to become true.
// Useful when multiple goroutines need to be woken up at once (Broadcast).

func syncCond() {

	var mu sync.Mutex
	cond := sync.NewCond(&mu)
	ready := false

	var wg sync.WaitGroup

	// Multiple waiters — all block until the condition is signaled
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			mu.Lock()
			for !ready { // always check the condition in a loop
				cond.Wait() // releases lock, sleeps, re-acquires lock on wake
			}
			mu.Unlock()

			fmt.Printf("  Waiter %d: proceeding!\n", id)
		}(i)
	}

	// Simulate setup work, then wake ALL waiters
	time.Sleep(50 * time.Millisecond)
	mu.Lock()
	ready = true
	mu.Unlock()

	cond.Broadcast() // wake all waiting goroutines (Signal wakes just one)
	wg.Wait()
	fmt.Println("  All waiters done!")
}

// Compare-and-Swap (CAS) — the building block of lock-free algorithms
// Only updates the value if it still matches the expected "old" value.

func compareAndSwap() {

	var state atomic.Int32
	state.Store(0) // 0=idle, 1=running, 2=done

	// Only one goroutine will win the CAS race to transition 0→1
	var wg sync.WaitGroup
	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			if state.CompareAndSwap(0, 1) {
				fmt.Printf("  Goroutine %d won the race! (0 → 1)\n", id)
				time.Sleep(20 * time.Millisecond)
				state.Store(2) // mark done
			} else {
				fmt.Printf("  Goroutine %d lost the race (state was already %d)\n", id, state.Load())
			}
		}(i)
	}
	wg.Wait()
	fmt.Printf("  Final state: %d\n", state.Load())
}

func main() {

	atomicCounter()
	atomicBool()
	atomicValue()
	syncMap()
	syncPool()
	syncCond()
	compareAndSwap()
}
