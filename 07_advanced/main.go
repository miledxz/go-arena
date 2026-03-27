package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// Context with cancellation
// context.WithCancel returns a ctx that can be cancelled manually.

func longRunningTask(ctx context.Context, id int, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("  Task %d: cancelled (%v)\n", id, ctx.Err())
			return
		default:
			fmt.Printf("  Task %d: working...\n", id)
			time.Sleep(50 * time.Millisecond)
		}
	}
}

func contextCancel() {

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go longRunningTask(ctx, i, &wg)
	}

	// Let them work for a bit, then cancel
	time.Sleep(130 * time.Millisecond)
	fmt.Println("  Cancelling all tasks...")
	cancel()

	wg.Wait()
	fmt.Println("  All tasks stopped!")
}

// Context with timeout
// Automatically cancels after a deadline.

func slowOperation(ctx context.Context) (string, error) {
	// Simulate work that takes 200ms
	select {
	case <-time.After(200 * time.Millisecond):
		return "operation complete", nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func contextTimeout() {

	// Timeout after 100ms – the 200ms operation will be cancelled
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel() // always call cancel to release resources

	result, err := slowOperation(ctx)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
	} else {
		fmt.Printf("  Result: %s\n", result)
	}

	// Now with a longer timeout – should succeed
	ctx2, cancel2 := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel2()

	result2, err2 := slowOperation(ctx2)
	if err2 != nil {
		fmt.Printf("  Error: %v\n", err2)
	} else {
		fmt.Printf("  Result: %s\n", result2)
	}
}

// Context with values

type contextKey string

const requestIDKey contextKey = "requestID"

func handleRequest(ctx context.Context) {
	reqID := ctx.Value(requestIDKey)
	fmt.Printf("  Handling request with ID: %v\n", reqID)
}

func contextValues() {

	ctx := context.WithValue(context.Background(), requestIDKey, "req-12345")
	handleRequest(ctx)

	ctx2 := context.WithValue(context.Background(), requestIDKey, "req-67890")
	handleRequest(ctx2)
}

// Rate Limiting
// Control how frequently operations happen using a ticker.

func rateLimiting() {

	// Allow 1 request every 50ms
	limiter := time.NewTicker(50 * time.Millisecond)
	defer limiter.Stop()

	requests := []string{"req-A", "req-B", "req-C", "req-D", "req-E"}

	start := time.Now()
	for _, req := range requests {
		<-limiter.C // wait for the next tick
		fmt.Printf("  Processing %s at %v\n", req, time.Since(start).Round(time.Millisecond))
	}
}

// Bursty Rate Limiting
// Allow bursts up to N, then rate limit.

func burstyRateLimiting() {

	// Allow bursts of 3, then 1 per 50ms
	burstyLimiter := make(chan time.Time, 3)

	// Pre-fill the burst capacity
	for i := 0; i < 3; i++ {
		burstyLimiter <- time.Now()
	}

	// Refill at a steady rate
	go func() {
		for t := range time.Tick(50 * time.Millisecond) {
			burstyLimiter <- t
		}
	}()

	requests := []string{"A", "B", "C", "D", "E", "F", "G"}
	start := time.Now()

	for _, req := range requests {
		<-burstyLimiter
		fmt.Printf("  Request %s at %v\n", req, time.Since(start).Round(time.Millisecond))
	}
}

// Semaphore Pattern
// Limit the number of concurrent goroutines using a buffered channel.

func semaphore() {

	const maxConcurrent = 3
	sem := make(chan struct{}, maxConcurrent) // semaphore

	var wg sync.WaitGroup

	for i := 1; i <= 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			sem <- struct{}{}        // acquire (blocks if 3 goroutines already running)
			defer func() { <-sem }() // release

			fmt.Printf("  Task %d: running (concurrent slots used: %d/%d)\n",
				id, len(sem), maxConcurrent)
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
		}(i)
	}

	wg.Wait()
	fmt.Println("  All tasks complete!")
}

// errgroup-like pattern (manual implementation)
// Run multiple goroutines and collect the first error.

func fetchURL(ctx context.Context, url string) error {
	// Simulate fetching – some URLs "fail"
	delay := time.Duration(rand.Intn(100)) * time.Millisecond

	select {
	case <-time.After(delay):
		if url == "https://fail.example.com" {
			return fmt.Errorf("failed to fetch %s", url)
		}
		fmt.Printf("  Fetched: %s\n", url)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func errgroup() {

	urls := []string{
		"https://api.example.com",
		"https://data.example.com",
		"https://fail.example.com",
		"https://cache.example.com",
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, len(urls))
	var wg sync.WaitGroup

	for _, url := range urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			if err := fetchURL(ctx, u); err != nil {
				errCh <- err
				cancel() // cancel remaining operations on first error
			}
		}(url)
	}

	wg.Wait()
	close(errCh)

	// Collect errors
	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		fmt.Printf("  Errors encountered: %v\n", errors.Join(errs...))
	} else {
		fmt.Println("  All fetches succeeded!")
	}
}

// Graceful shutdown pattern
// Coordinate shutdown of multiple components.

type Server struct {
	name string
}

func (s *Server) Start(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Printf("  [%s] Server started\n", s.name)

	ticker := time.NewTicker(40 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fmt.Printf("  [%s] Handling request...\n", s.name)
		case <-ctx.Done():
			fmt.Printf("  [%s] Shutting down gracefully...\n", s.name)
			// Simulate cleanup
			time.Sleep(20 * time.Millisecond)
			fmt.Printf("  [%s] Shutdown complete\n", s.name)
			return
		}
	}
}

func gracefulShutdown() {

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	servers := []*Server{
		{name: "HTTP"},
		{name: "gRPC"},
		{name: "Worker"},
	}

	for _, srv := range servers {
		wg.Add(1)
		go srv.Start(ctx, &wg)
	}

	// Simulate running for a while, then shutdown
	time.Sleep(150 * time.Millisecond)
	fmt.Println("\n  === INITIATING SHUTDOWN ===")
	cancel()

	wg.Wait()
	fmt.Println("  All servers shut down cleanly!")
}

func main() {

	contextCancel()
	contextTimeout()
	contextValues()
	rateLimiting()
	burstyRateLimiting()
	semaphore()
	errgroup()
	gracefulShutdown()
}
