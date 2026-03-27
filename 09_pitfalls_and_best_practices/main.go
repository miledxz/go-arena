package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// PITFALL 1: Goroutine leak — goroutine that never exits
// A leaked goroutine consumes memory and CPU forever.
// Rule: always know when and how every goroutine will exit.

func goroutineLeak_BAD() {

	// BAD: this goroutine blocks forever because nobody reads from ch
	ch := make(chan int)
	go func() {
		// This goroutine will NEVER exit because the send blocks forever.
		// In real code, this is a memory leak.
		result := 42
		select {
		case ch <- result:
		// If we don't drain ch, this goroutine is stuck forever.
		default:
			fmt.Println("  [LEAK] goroutine could not send — would leak in real code")
		}
	}()

	time.Sleep(50 * time.Millisecond)
	fmt.Println("  Leaked goroutine is stuck, waiting to send on a channel nobody reads")
}

// FIX: always provide a way out — use context or done channel.
func goroutineLeak_FIXED() {
	fmt.Println("\n--- Fix: Cancel with context ---")

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	ch := make(chan int)
	go func() {
		select {
		case ch <- 42:
			fmt.Println("  Sent result")
		case <-ctx.Done():
			fmt.Println("  [SAFE] goroutine exited cleanly via context")
		}
	}()

	// We intentionally don't read from ch — but the goroutine still exits cleanly
	time.Sleep(100 * time.Millisecond)
}

// PITFALL 2: Deadlock — all goroutines are asleep
// Go detects when ALL goroutines are blocked and panics with "deadlock".
// In real apps with many goroutines, partial deadlocks are silent and dangerous.

func deadlockExample() {

	// Pattern A: Sending on unbuffered channel with no receiver
	// ch := make(chan int)
	// ch <- 1  // DEADLOCK: main goroutine blocks, nobody to receive
	fmt.Println("  Pattern A: send on unbuffered channel with no receiver = deadlock")

	// Pattern B: Two goroutines waiting on each other
	// ch1 := make(chan int)
	// ch2 := make(chan int)
	// go func() { <-ch1; ch2 <- 1 }()
	// go func() { <-ch2; ch1 <- 1 }()
	// Both wait for the other to send first = deadlock
	fmt.Println("  Pattern B: two goroutines waiting on each other = deadlock")

	// Pattern C: Forgetting to close a channel that range loops over
	// ch := make(chan int)
	// go func() { ch <- 1; ch <- 2 }()  // forgot close(ch)
	// for v := range ch { ... }  // blocks forever after receiving 2 values
	fmt.Println("  Pattern C: range over channel that is never closed = hangs forever")

	// SAFE pattern: always close when done
	ch := make(chan int)
	go func() {
		ch <- 1
		ch <- 2
		close(ch) // range loop will exit after this
	}()
	for v := range ch {
		fmt.Printf("  Received: %d\n", v)
	}
	fmt.Println("  Safe: channel properly closed, range exits cleanly")
}

// PITFALL 3: Closure variable capture
// Go 1.22+ fixed this for `for` loops, but it's still important to understand.
// In older Go versions, all goroutines share the same loop variable.

func closureCapture() {

	// In Go < 1.22 this prints "5" five times because all goroutines
	// share the same `i` variable. In Go 1.22+ the loop variable
	// is per-iteration, so it works correctly. But passing as a param
	// is still the safest and most portable approach.

	fmt.Println("  Safe approach — pass as parameter:")
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			fmt.Printf("    n=%d\n", n) // always correct
		}(i)
	}
	wg.Wait()
}

// PITFALL 4: Channels don't broadcast — each value goes to ONE receiver
// If you need to notify multiple goroutines, use close() or sync.Cond.

func channelNotBroadcast() {

	ch := make(chan string, 1)

	var wg sync.WaitGroup
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			select {
			case msg := <-ch:
				fmt.Printf("  Receiver %d got: %s\n", id, msg)
			case <-time.After(50 * time.Millisecond):
				fmt.Printf("  Receiver %d got nothing\n", id)
			}
		}(i)
	}

	ch <- "hello" // only ONE of the 3 receivers gets this
	wg.Wait()

	fmt.Println("\n  FIX: Use close() to broadcast a signal to ALL goroutines:")
	done := make(chan struct{})
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			<-done // ALL goroutines unblock when done is closed
			fmt.Printf("  Receiver %d: got broadcast signal!\n", id)
		}(i)
	}
	close(done) // wakes up ALL waiting goroutines at once
	wg.Wait()
}

// PITFALL 5: Sending on a closed channel panics
// Only the sender should close a channel. Never close from the receiver side.

func sendOnClosedChannel() {

	ch := make(chan int, 1)
	close(ch)

	// ch <- 1  // PANIC: send on closed channel

	// Safe: check if you should send using a done/ctx signal
	fmt.Println("  Rule: only the SENDER closes the channel")
	fmt.Println("  Rule: never close a channel from the receiver side")
	fmt.Println("  Rule: closing an already-closed channel also panics")

	// Reading from a closed channel returns the zero value
	val, ok := <-ch
	fmt.Printf("  Read from closed channel: val=%d, ok=%v\n", val, ok)
}

// BEST PRACTICE: Generator pattern
// A function returns a read-only channel. The goroutine owns the channel
// and closes it when done. Clean, composable, and leak-free.

func fibonacci(n int) <-chan int {
	ch := make(chan int)
	go func() {
		defer close(ch)
		a, b := 0, 1
		for i := 0; i < n; i++ {
			ch <- a
			a, b = b, a+b
		}
	}()
	return ch
}

func generatorPattern() {

	fmt.Print("  Fibonacci: ")
	for val := range fibonacci(10) {
		fmt.Printf("%d ", val)
	}
	fmt.Println()
}

// BEST PRACTICE: Tee channel
// Split one input channel into two output channels.
// Both outputs receive every value.

func tee(done <-chan struct{}, in <-chan int) (<-chan int, <-chan int) {
	out1 := make(chan int)
	out2 := make(chan int)

	go func() {
		defer close(out1)
		defer close(out2)
		for val := range in {
			// Use local copies so both sends can proceed independently
			o1, o2 := out1, out2
			for i := 0; i < 2; i++ {
				select {
				case o1 <- val:
					o1 = nil // disable this case after sending
				case o2 <- val:
					o2 = nil
				case <-done:
					return
				}
			}
		}
	}()
	return out1, out2
}

func teeChannel() {
	fmt.Println("\n--- Best Practice: Tee Channel ---")

	done := make(chan struct{})
	defer close(done)

	// Source channel
	source := make(chan int)
	go func() {
		defer close(source)
		for i := 1; i <= 4; i++ {
			source <- i
		}
	}()

	out1, out2 := tee(done, source)

	// Read from both outputs concurrently
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for v := range out1 {
			fmt.Printf("  out1: %d\n", v)
		}
	}()
	go func() {
		defer wg.Done()
		for v := range out2 {
			fmt.Printf("  out2: %d\n", v)
		}
	}()
	wg.Wait()
}

// BEST PRACTICE: Bridge channel
// Flatten a channel-of-channels into a single channel.
// Useful when you have a sequence of channels produced dynamically.

func bridge(done <-chan struct{}, chanOfChans <-chan <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for ch := range chanOfChans {
			for val := range ch {
				select {
				case out <- val:
				case <-done:
					return
				}
			}
		}
	}()
	return out
}

func bridgeChannel() {

	done := make(chan struct{})
	defer close(done)

	// Create a channel of channels
	chanOfChans := make(chan (<-chan int))
	go func() {
		defer close(chanOfChans)
		for i := 0; i < 3; i++ {
			ch := make(chan int, 2)
			ch <- i * 10
			ch <- i*10 + 1
			close(ch)
			chanOfChans <- ch
		}
	}()

	// Bridge flattens them into one stream
	for val := range bridge(done, chanOfChans) {
		fmt.Printf("  Bridged: %d\n", val)
	}
}

// BEST PRACTICE: Bounded parallelism
// Process items concurrently but limit to N at a time.
// Combines a semaphore with WaitGroup for clean bounded work.

func processItem(id int) string {
	time.Sleep(30 * time.Millisecond) // simulate work
	return fmt.Sprintf("result-%d", id)
}

func boundedParallelism() {

	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	maxConcurrent := 3

	results := make([]string, len(items))
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	for i, item := range items {
		wg.Add(1)
		sem <- struct{}{} // acquire slot

		go func(idx, val int) {
			defer wg.Done()
			defer func() { <-sem }() // release slot

			results[idx] = processItem(val) // each goroutine writes its own index — safe
		}(i, item)
	}
	wg.Wait()

	fmt.Println("  Results:")
	for _, r := range results {
		fmt.Printf("    %s\n", r)
	}
}

func main() {

	// Pitfalls
	goroutineLeak_BAD()
	goroutineLeak_FIXED()
	deadlockExample()
	closureCapture()
	channelNotBroadcast()
	sendOnClosedChannel()

	// Best practices & patterns
	generatorPattern()
	teeChannel()
	bridgeChannel()
	boundedParallelism()
}
