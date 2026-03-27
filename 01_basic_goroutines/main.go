package main

import (
	"fmt"
	"time"
)

func sayHello(name string) {
	fmt.Printf("Hello, %s!\n", name)
}

func firstGoroutine() {

	go sayHello("World")

	// Without this sleep, main() might exit before the goroutine finishes.
	// We'll learn better ways to wait later (WaitGroups, channels).
	time.Sleep(100 * time.Millisecond)
}

// Multiple goroutines
// Each goroutine runs concurrently; order is NOT guaranteed.
func multipleGoroutines() {

	for i := 1; i <= 5; i++ {
		go func(id int) {
			fmt.Printf("  Goroutine %d is running\n", id)
		}(i) // we pass `i` as an argument to avoid closure pitfalls
	}

	time.Sleep(200 * time.Millisecond)
}

func manyGoroutines() {

	start := time.Now()
	for i := 0; i < 1000; i++ {
		go func(id int) {
			// simulate a tiny bit of work
			_ = id * id
		}(i)
	}

	time.Sleep(200 * time.Millisecond)
	fmt.Printf("  Launched 1000 goroutines in %v\n", time.Since(start))
}

func anonymousGoroutine() {

	go func() {
		fmt.Println("  I'm running inside an anonymous goroutine!")
	}()

	time.Sleep(100 * time.Millisecond)
}

func periodicGoroutine() {

	go func() {
		for i := 1; i <= 3; i++ {
			fmt.Printf("  Tick #%d\n", i)
			time.Sleep(50 * time.Millisecond)
		}
	}()

	// Wait long enough for the goroutine to finish
	time.Sleep(200 * time.Millisecond)
	fmt.Println("  Done ticking!")
}

func main() {

	firstGoroutine()
	multipleGoroutines()
	manyGoroutines()
	anonymousGoroutine()
	periodicGoroutine()
}
