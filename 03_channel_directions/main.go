package main

import "fmt"

// Directional channels in function signatures

// producer can only SEND to the channel
func producer(out chan<- int) {
	for i := 1; i <= 5; i++ {
		out <- i
	}
	close(out)
}

// consumer can only RECEIVE from the channel
func consumer(in <-chan int) {
	for val := range in {
		fmt.Printf("  Consumed: %d\n", val)
	}
}

func directionalChannels() {

	ch := make(chan int)
	go producer(ch)
	consumer(ch) // blocks until producer closes the channel
}

// The comma-ok idiom for checking if a channel is closed
func commaOk() {

	ch := make(chan int, 3)
	ch <- 10
	ch <- 20
	close(ch)

	// Using the comma-ok pattern:
	for {
		val, ok := <-ch
		if !ok {
			fmt.Println("  Channel closed!")
			break
		}
		fmt.Printf("  Received: %d (ok=%v)\n", val, ok)
	}
}

// Chaining goroutines with channels (pipeline stage)
// Each function takes input from one channel and writes output to another.

func generateNumbers(out chan<- int) {
	for i := 1; i <= 5; i++ {
		out <- i
	}
	close(out)
}

func doubleNumbers(in <-chan int, out chan<- int) {
	for val := range in {
		out <- val * 2
	}
	close(out)
}

func printNumbers(in <-chan int) {
	for val := range in {
		fmt.Printf("  %d\n", val)
	}
}

func chainedGoroutines() {

	// Stage 1 → Stage 2 → Print
	stage1 := make(chan int)
	stage2 := make(chan int)

	go generateNumbers(stage1)
	go doubleNumbers(stage1, stage2)
	printNumbers(stage2)
}

// done channel pattern (signal without data)
// Using an empty struct chan struct{} is idiomatic – it uses zero memory.
func worker(done chan<- struct{}) {
	fmt.Println("  Worker: doing important work...")
	fmt.Println("  Worker: finished!")
	done <- struct{}{} // signal done
}

func doneChannel() {

	done := make(chan struct{})
	go worker(done)
	<-done // wait for the signal
	fmt.Println("  Main: worker is done!")
}

// Collecting results from multiple goroutines
func square(n int, results chan<- string) {
	results <- fmt.Sprintf("  %d² = %d", n, n*n)
}

func collectResults() {

	results := make(chan string, 5)

	nums := []int{2, 4, 6, 8, 10}
	for _, n := range nums {
		go square(n, results)
	}

	// Collect all results (we know exactly how many to expect)
	for i := 0; i < len(nums); i++ {
		fmt.Println(<-results)
	}
}

func main() {

	directionalChannels()
	commaOk()
	chainedGoroutines()
	doneChannel()
	collectResults()
}
