# go-arena

Learning Go concurrency by doing. Each folder is a standalone program you can run and tinker with.

```
01_basic_goroutines        — launching goroutines, closures
02_channels_basics         — send, receive, buffered channels
03_channel_directions      — directional channels, pipelines, done signal
04_select_statement        — select, timeouts, tickers, non-blocking ops
05_sync_primitives         — WaitGroup, Mutex, RWMutex, sync.Once
06_patterns                — worker pool, fan-in, fan-out, pipelines, or-done
07_advanced                — context, rate limiting, semaphore, errgroup, graceful shutdown
08_atomics_and_sync_tools  — sync/atomic, sync.Map, sync.Pool, sync.Cond, CAS
09_pitfalls_and_best_practices — goroutine leaks, deadlocks, generator, tee, bridge
```

Pick a folder and run it:

```bash
go run ./01_basic_goroutines/
```
