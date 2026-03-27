package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// Worker Pool
// A fixed number of workers process jobs from a shared channel.
// This limits concurrency and prevents overwhelming resources.

type Job struct {
	ID      int
	Payload int
}

type Result struct {
	JobID  int
	Output int
}

func workerPoolWorker(id int, jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		// Simulate work
		time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
		results <- Result{
			JobID:  job.ID,
			Output: job.Payload * job.Payload, // compute square
		}
		fmt.Printf("  Worker %d processed job %d\n", id, job.ID)
	}
}

func workerPoolPattern() {

	const numWorkers = 3
	const numJobs = 10

	jobs := make(chan Job, numJobs)
	results := make(chan Result, numJobs)

	// Start workers
	var wg sync.WaitGroup
	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go workerPoolWorker(w, jobs, results, &wg)
	}

	// Send jobs
	for j := 1; j <= numJobs; j++ {
		jobs <- Job{ID: j, Payload: j}
	}
	close(jobs) // no more jobs

	// Wait for all workers to finish, then close results
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	fmt.Println("\n  Results:")
	for res := range results {
		fmt.Printf("    Job %d → %d\n", res.JobID, res.Output)
	}
}

// Fan-Out
// Distribute work from one source to multiple goroutines.

func fanOutWorker(id int, input <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	for val := range input {
		fmt.Printf("  Fan-out worker %d got: %d\n", id, val)
		time.Sleep(20 * time.Millisecond) // simulate work
	}
}

func fanOutPattern() {

	input := make(chan int, 10)

	// Send data
	go func() {
		for i := 1; i <= 9; i++ {
			input <- i
		}
		close(input)
	}()

	// Fan out to 3 workers
	var wg sync.WaitGroup
	for w := 1; w <= 3; w++ {
		wg.Add(1)
		go fanOutWorker(w, input, &wg)
	}
	wg.Wait()
}

// Fan-In
// Merge multiple input channels into a single output channel.

func fanInSource(name string, count int) <-chan string {
	ch := make(chan string)
	go func() {
		defer close(ch)
		for i := 1; i <= count; i++ {
			ch <- fmt.Sprintf("%s-%d", name, i)
			time.Sleep(time.Duration(rand.Intn(30)) * time.Millisecond)
		}
	}()
	return ch
}

// fanIn merges multiple channels into one
func fanIn(channels ...<-chan string) <-chan string {
	merged := make(chan string)
	var wg sync.WaitGroup

	for _, ch := range channels {
		wg.Add(1)
		go func(c <-chan string) {
			defer wg.Done()
			for val := range c {
				merged <- val
			}
		}(ch)
	}

	go func() {
		wg.Wait()
		close(merged)
	}()

	return merged
}

func fanInPattern() {

	// Three independent sources
	src1 := fanInSource("alpha", 3)
	src2 := fanInSource("beta", 3)
	src3 := fanInSource("gamma", 3)

	// Merge them all into one
	merged := fanIn(src1, src2, src3)

	for msg := range merged {
		fmt.Printf("  Received: %s\n", msg)
	}
}

// Pipeline
// Each stage reads from input, transforms, writes to output.
// Compose stages like Unix pipes.

// Stage: generate numbers
func generate(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			out <- n
		}
	}()
	return out
}

// Stage: square each number
func squareStage(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			out <- n * n
		}
	}()
	return out
}

// Stage: filter – only pass even numbers
func filterEven(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			if n%2 == 0 {
				out <- n
			}
		}
	}()
	return out
}

// Stage: add a constant
func addN(in <-chan int, n int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for val := range in {
			out <- val + n
		}
	}()
	return out
}

func pipelinePattern() {

	// Pipeline: generate → square → filterEven → add(10) → print
	nums := generate(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
	squared := squareStage(nums)
	evens := filterEven(squared)
	final := addN(evens, 10)

	fmt.Println("  Pipeline: generate → square → filterEven → add(10)")
	for val := range final {
		fmt.Printf("    → %d\n", val)
	}
}

// Or-Done Channel
// Read from a channel but respect a done signal.

func orDone(done <-chan struct{}, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for {
			select {
			case <-done:
				return
			case val, ok := <-in:
				if !ok {
					return
				}
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

func orDonePattern() {

	// Create an infinite stream
	stream := make(chan int)
	go func() {
		i := 0
		for {
			stream <- i
			i++
			time.Sleep(10 * time.Millisecond)
		}
	}()

	done := make(chan struct{})

	// Cancel after getting 5 items
	go func() {
		time.Sleep(60 * time.Millisecond)
		close(done)
	}()

	for val := range orDone(done, stream) {
		fmt.Printf("  Got: %d\n", val)
	}
	fmt.Println("  Stream cancelled via done channel")
}

func main() {

	workerPoolPattern()
	fanOutPattern()
	fanInPattern()
	pipelinePattern()
	orDonePattern()
}
