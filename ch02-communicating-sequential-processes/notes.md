# Chapter 2
## The difference between concurrency and parallelism

### What is concurrency
Concurrency is a property of *code structure*. A program is concurrent when it is
written to handle multiple things at once, independent tasks that can be
composed and interleaved. It says nothing about what happens at runtime.

A Go program with two goroutines updating a counter is concurrent regardless of
the hardware it runs on.

### What is parallelism
Parallelism is a property of *execution*. Two tasks run in parallel when they
are literally executing at the same instant on separate CPU cores.

Parallelism requires both concurrent code *and* hardware with multiple cores
available. In Go, `GOMAXPROCS` controls how many OS threads (and therefore cores)
the runtime may use. On a single-core machine, concurrent goroutines interleave,
they never run in parallel.

**The key distinction:** you write concurrency into your program, the runtime and
hardware determine whether parallelism happens. Concurrency is the design.
Parallelism is an emergent property.

### What is CSP
CSP (Communicating Sequential Processes) is a 1978 paper by Tony Hoare proposing
that communication and synchronisation between processes should be *first-class
primitives* in a language, not bolted on via shared memory and locks.

Most languages stop their concurrency abstraction at the OS thread level, leaving
synchronisation to the programmer (mutexes, semaphores). Go takes a different
route: it supplants OS threads with goroutines, and makes channel-based
communication the idiomatic way to coordinate between them.

This is where Go's core philosophy comes from:
> "Don't communicate by sharing memory; share memory by communicating."

Channels are Go's expression of CSP's communication primitives. Instead of two
goroutines reaching into shared state, they send values to each other — and the
act of sending is itself the synchronization point.