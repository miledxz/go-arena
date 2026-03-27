package main

import (
	"fmt"
	"sync"
	"time"
)

// WaitGroup – wait for goroutines to finish
func waitGroup() {
	fmt.Println("\n--- Example 1: WaitGroup ---")

	var wg sync.WaitGroup

	for i := 1; i <= 5; i++ {
		wg.Add(1) // increment the counter BEFORE launching the goroutine

		go func(id int) {
			defer wg.Done() // decrement when this goroutine finishes
			fmt.Printf("  Worker %d: starting\n", id)
			time.Sleep(time.Duration(id*20) * time.Millisecond)
			fmt.Printf("  Worker %d: done\n", id)
		}(i)
	}

	wg.Wait() // block until counter reaches zero
	fmt.Println("  All workers finished!")
}

// Race condition WITHOUT mutex (DANGER!)
// Multiple goroutines modify a shared variable – result is unpredictable.
func raceCondition() {
	fmt.Println("\n--- Example 2: Race Condition (unsafe!) ---")

	counter := 0
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter++ // ← NOT SAFE! Multiple goroutines read+write concurrently
		}()
	}

	wg.Wait()
	// Expected: 1000, Actual: often less due to race condition
	fmt.Printf("  Counter (unsafe): %d (expected 1000)\n", counter)
}

// Fix the race condition WITH Mutex
func mutex() {
	fmt.Println("\n--- Example 3: Mutex (safe!) ---")

	counter := 0
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()   // acquire the lock
			counter++   // safe: only one goroutine can be here at a time
			mu.Unlock() // release the lock
		}()
	}

	wg.Wait()
	fmt.Printf("  Counter (safe): %d (expected 1000)\n", counter)
}

// sync.Once – initialize exactly once
// Even with multiple goroutines calling it, the function runs only once.
func syncOnce() {
	fmt.Println("\n--- Example 4: sync.Once ---")

	var once sync.Once
	var wg sync.WaitGroup

	initFunc := func() {
		fmt.Println("  Initializing... (this prints only once)")
	}

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			fmt.Printf("  Goroutine %d: calling once.Do\n", id)
			once.Do(initFunc) // only the first call actually runs initFunc
		}(i)
	}

	wg.Wait()
}

// RWMutex – multiple concurrent readers, exclusive writer
// Great for read-heavy workloads.
type SafeMap struct {
	mu   sync.RWMutex
	data map[string]int
}

func NewSafeMap() *SafeMap {
	return &SafeMap{data: make(map[string]int)}
}

func (sm *SafeMap) Get(key string) (int, bool) {
	sm.mu.RLock() // read lock – multiple readers allowed
	defer sm.mu.RUnlock()
	val, ok := sm.data[key]
	return val, ok
}

func (sm *SafeMap) Set(key string, val int) {
	sm.mu.Lock() // write lock – exclusive access
	defer sm.mu.Unlock()
	sm.data[key] = val
}

func rwMutex() {
	fmt.Println("\n--- Example 5: RWMutex (Safe Map) ---")

	sm := NewSafeMap()
	var wg sync.WaitGroup

	// Writers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", i)
			sm.Set(key, i*10)
			fmt.Printf("  Writer: set %s = %d\n", key, i*10)
		}(i)
	}

	// Wait for writers to finish, then read
	wg.Wait()

	// Readers (safe to run concurrently)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", i)
			if val, ok := sm.Get(key); ok {
				fmt.Printf("  Reader: %s = %d\n", key, val)
			}
		}(i)
	}

	wg.Wait()
}

// WaitGroup + Mutex combined – concurrent word counter
func wordCounter() {

	sentences := []string{
		"the quick brown fox",
		"jumps over the working dog",
		"the fox is quick and the dog is working",
	}

	wordCount := make(map[string]int)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, sentence := range sentences {
		wg.Add(1)
		go func(s string) {
			defer wg.Done()
			// Count words in this sentence
			word := ""
			for _, ch := range s + " " {
				if ch == ' ' {
					if word != "" {
						mu.Lock()
						wordCount[word]++
						mu.Unlock()
						word = ""
					}
				} else {
					word += string(ch)
				}
			}
		}(sentence)
	}

	wg.Wait()

	fmt.Println("  Word counts:")
	for word, count := range wordCount {
		fmt.Printf("    %-8s: %d\n", word, count)
	}
}

func main() {

	waitGroup()
	raceCondition()
	mutex()
	syncOnce()
	rwMutex()
	wordCounter()
}
