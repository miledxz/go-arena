package main

import (
	"fmt"
	"time"
)

// Basic select – wait on two channels
func basicSelect() {

	ch1 := make(chan string)
	ch2 := make(chan string)

	go func() {
		time.Sleep(50 * time.Millisecond)
		ch1 <- "message from channel 1"
	}()

	go func() {
		time.Sleep(30 * time.Millisecond)
		ch2 <- "message from channel 2"
	}()

	// Wait for EITHER channel – whichever is ready first wins
	select {
	case msg := <-ch1:
		fmt.Println("  Received:", msg)
	case msg := <-ch2:
		fmt.Println("  Received:", msg)
	}
}

// Timeout with select
// Prevent waiting forever by using time.After
func timeout() {

	ch := make(chan string)

	go func() {
		time.Sleep(200 * time.Millisecond) // simulate slow work
		ch <- "result"
	}()

	select {
	case msg := <-ch:
		fmt.Println("  Got result:", msg)
	case <-time.After(100 * time.Millisecond):
		fmt.Println("  Timed out! No result within 100ms")
	}
}

// Non-blocking operations with default
func nonBlocking() {

	ch := make(chan int, 1)

	// Non-blocking receive
	select {
	case val := <-ch:
		fmt.Println("  Received:", val)
	default:
		fmt.Println("  No value ready, moving on...")
	}

	// Non-blocking send
	ch <- 42
	select {
	case ch <- 99:
		fmt.Println("  Sent 99")
	default:
		fmt.Println("  Channel full, couldn't send")
	}

	fmt.Println("  Value in channel:", <-ch)
}

// Select in a loop – multiplexing multiple sources
// ----------------------------------------------------------------------------
func selectLoop() {

	fast := make(chan string)
	slow := make(chan string)
	quit := make(chan struct{})

	// Fast producer
	go func() {
		for i := 0; i < 3; i++ {
			time.Sleep(30 * time.Millisecond)
			fast <- fmt.Sprintf("fast-%d", i)
		}
	}()

	// Slow producer
	go func() {
		for i := 0; i < 2; i++ {
			time.Sleep(70 * time.Millisecond)
			slow <- fmt.Sprintf("slow-%d", i)
		}
	}()

	// Quit after a while
	go func() {
		time.Sleep(250 * time.Millisecond)
		close(quit)
	}()

	// Multiplex
	for {
		select {
		case msg := <-fast:
			fmt.Println("  [FAST]", msg)
		case msg := <-slow:
			fmt.Println("  [SLOW]", msg)
		case <-quit:
			fmt.Println("  Quit signal received!")
			return
		}
	}
}

// Ticker + Done pattern (very common in real code)
func tickerDone() {

	ticker := time.NewTicker(50 * time.Millisecond)
	done := make(chan struct{})

	go func() {
		time.Sleep(275 * time.Millisecond)
		close(done)
	}()

	count := 0
	for {
		select {
		case t := <-ticker.C:
			count++
			fmt.Printf("  Tick #%d at %v\n", count, t.Format("15:04:05.000"))
		case <-done:
			ticker.Stop()
			fmt.Println("  Ticker stopped!")
			return
		}
	}
}

// Random selection when multiple cases are ready
func randomSelection() {

	ch1 := make(chan string, 1)
	ch2 := make(chan string, 1)

	// Both channels are ready simultaneously
	ch1 <- "from ch1"
	ch2 <- "from ch2"

	counts := map[string]int{"ch1": 0, "ch2": 0}

	for i := 0; i < 100; i++ {
		// Refill both channels
		if len(ch1) == 0 {
			ch1 <- "from ch1"
		}
		if len(ch2) == 0 {
			ch2 <- "from ch2"
		}

		select {
		case <-ch1:
			counts["ch1"]++
		case <-ch2:
			counts["ch2"]++
		}
	}
	fmt.Printf("  ch1 selected %d times, ch2 selected %d times\n", counts["ch1"], counts["ch2"])
	fmt.Println("  (Should be roughly 50/50 – Go picks randomly!)")
}

func main() {

	basicSelect()
	timeout()
	nonBlocking()
	selectLoop()
	tickerDone()
	randomSelection()
}
