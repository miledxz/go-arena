package main

import "fmt"

func basicChannel() {
	// Create an unbuffered channel of type string
	ch := make(chan string)

	// Sender goroutine
	go func() {
		ch <- "Hello from the goroutine!" // send a value into the channel
	}()

	// Receive the value (blocks until the goroutine sends)
	msg := <-ch
	fmt.Println("  Received:", msg)
}

// Channel as a synchronization tool
// No need for time.Sleep – the channel receive waits for the goroutine.
func syncWithChannel() {

	done := make(chan bool)

	go func() {
		fmt.Println("  Working...")
		fmt.Println("  Done working!")
		done <- true // signal completion
	}()

	<-done // wait until the goroutine signals
	fmt.Println("  Main received done signal")
}

// Sending multiple values through a channel
func multipleValues() {

	ch := make(chan int)

	go func() {
		for i := 1; i <= 5; i++ {
			ch <- i * i // send squares
		}
		close(ch) // close the channel when done sending
	}()

	// Range over channel – automatically stops when channel is closed
	for val := range ch {
		fmt.Printf("  Received: %d\n", val)
	}
}

// Buffered channels
// A buffered channel can hold N values without a receiver being ready.
func bufferedChannel() {
	// Buffer size of 3: can hold 3 values without blocking
	ch := make(chan string, 3)

	ch <- "first"
	ch <- "second"
	ch <- "third"
	// ch <- "fourth"  // ← This WOULD block (buffer full, no receiver)

	fmt.Println("  Buffer length:", len(ch), "/ capacity:", cap(ch))

	// Receive values
	fmt.Println(" ", <-ch)
	fmt.Println(" ", <-ch)
	fmt.Println(" ", <-ch)
}

// Passing data between goroutines through a channel
// One goroutine generates data, another consumes it.
func producerConsumer() {

	ch := make(chan string)

	// Producer goroutine
	go func() {
		fruits := []string{"apple", "banana", "cherry", "date"}
		for _, fruit := range fruits {
			ch <- fruit
		}
		close(ch)
	}()

	// Consumer (main goroutine)
	for fruit := range ch {
		fmt.Printf("  Consumed: %s\n", fruit)
	}
}

func computeSum(nums []int, resultCh chan int) {
	sum := 0
	for _, n := range nums {
		sum += n
	}
	resultCh <- sum
}

func returnResult() {

	nums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	ch := make(chan int)

	// Split work: first half + second half
	mid := len(nums) / 2
	go computeSum(nums[:mid], ch)
	go computeSum(nums[mid:], ch)

	// Collect results
	sum1 := <-ch
	sum2 := <-ch
	fmt.Printf("  Sum of halves: %d + %d = %d\n", sum1, sum2, sum1+sum2)
}

func main() {

	basicChannel()
	syncWithChannel()
	multipleValues()
	bufferedChannel()
	producerConsumer()
	returnResult()
}
