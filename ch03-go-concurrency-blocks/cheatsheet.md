# Concurrency in Go — Chapter 3 Study Summary

## 1. Goroutines

### What they are
- Functions running **concurrently** (not necessarily in parallel).
- Higher abstraction than OS threads or green threads → **coroutines**.
- **Non-preemptive**: can't be interrupted at arbitrary points.
- Go runtime automatically suspends them when they block and resumes them when unblocked.
- Scheduler is **M:N**: maps M green threads to N OS threads.
- Follow the **fork-join** model: a goroutine "forks" off, and at the "join point" it rejoins the parent.
- Stack starts at a few KB and grows/shrinks automatically.

### Syntax
```go
go sayHello()                    // named function

go func() {                      // anonymous (must invoke immediately)
    fmt.Println("hello")
}()

sayHello := func() { ... }       // assigned, then launched
go sayHello()
```

### Critical gotchas
- **Closures capture variables by reference**, not by value.
- **Loop variable trap**: classic bug — the loop usually finishes before goroutines run, so they all see the last value. The variable is moved to the heap because goroutines still hold a reference.
- **Fix**: pass the loop variable as a parameter to the closure so a copy is made:
  ```go
  for _, s := range salutations {
      wg.Add(1)
      go func(s string) {        // s is now a copy
          defer wg.Done()
          fmt.Println(s)
      }(s)
  }
  ```

---

## 2. The `sync` Package

### `sync.WaitGroup`
**Purpose:** Wait for a set of goroutines to finish when you don't care about their return values.

| Method | What it does |
|--------|--------------|
| `Add(delta int)` | Increment counter by `delta` (N goroutines starting). |
| `Done()` | Decrement counter by 1. Always use with `defer`. |
| `Wait()` | Block until counter reaches 0. |

**Tips:**
- Call `Add()` **outside** the goroutine — otherwise race condition.
- Place `Add()` calls as close as possible to the goroutine they track.
- If you need results, use channels + `select` instead.

---

### `sync.Mutex`
**Purpose:** Mutual exclusion to guard a critical section (shared resource).

| Method | What it does |
|--------|--------------|
| `Lock()` | Acquire exclusive access. |
| `Unlock()` | Release. Always `defer` it. |

**Tips:**
- Critical sections are expensive — keep them minimal.
- Always `defer lock.Unlock()` so it runs even on panic (otherwise → deadlock).

---

### `sync.RWMutex`
**Purpose:** Same as Mutex, but distinguishes readers from writers — many readers OR one writer.

| Method | What it does |
|--------|--------------|
| `Lock()` / `Unlock()` | Writer (exclusive). |
| `RLock()` / `RUnlock()` | Reader (shared with other readers). |

**Tip:** Prefer `RWMutex` over `Mutex` whenever reads dominate writes.

---

### `sync.Cond`
**Purpose:** Rendezvous point where goroutines wait for or announce an event.

**Constructor:** `sync.NewCond(l sync.Locker)` — takes anything implementing `sync.Locker`.

| Method | What it does |
|--------|--------------|
| `Wait()` | Suspends the goroutine. **Releases the Locker on entry, re-acquires it on exit.** |
| `Signal()` | Wakes **one** waiting goroutine (FIFO — longest waiter first). |
| `Broadcast()` | Wakes **all** waiting goroutines. |

**Canonical pattern:**
```go
c.L.Lock()
for !conditionTrue() {        // ALWAYS a `for`, never `if`
    c.Wait()
}
// ... critical work ...
c.L.Unlock()
```

**Tip:** `Wait()` doesn't hold the lock the entire time — it releases on entry, reacquires on exit.

---

### `sync.Once`
**Purpose:** Execute a function **exactly once**, ever, across all goroutines.

| Method | What it does |
|--------|--------------|
| `Do(f func())` | Runs `f` only on the first call; subsequent calls are no-ops. |

**Tip:** ⚠️ Will **deadlock** if `f` calls another function that's also gated by the same `Once`.

---

### `sync.Pool`
**Purpose:** Concurrent-safe object pool — constrain the creation of expensive objects (e.g., DB connections) while still serving many request callers.

| Member | What it does |
|--------|--------------|
| `Get() interface{}` | Returns an existing instance if available, else calls `New`. |
| `Put(x interface{})` | Returns the instance to the pool for reuse. |
| `New func() interface{}` | Factory function for new instances (must be thread-safe). |

**Why use it:** Reuse objects instead of letting the GC churn them; also useful for pre-warming caches.

**Rules:**
1. `New` must be thread-safe.
2. Make **no assumptions** about the state of an object you `Get`.
3. Always `Put` when done (use `defer`), or the pool is useless.
4. Objects in the pool should be **roughly uniform**.

---

## 3. Channels

### Declaring & instantiating
```go
var c chan T          // bidirectional
var c <-chan T        // receive-only
var c chan<- T        // send-only

c := make(chan T)     // unbuffered (capacity 0)
c := make(chan T, 4)  // buffered, capacity 4
```

Go implicitly converts bidirectional → unidirectional (not the reverse). Use this in function signatures to enforce intent.

### Operations
| Operation | Syntax |
|-----------|--------|
| Send | `c <- v` |
| Receive | `v := <-c` |
| Receive + check | `v, ok := <-c` — `ok=false` means closed **and** drained |
| Close | `close(c)` |
| Range | `for v := range c { ... }` — exits automatically on close |

### Blocking rules — the operations table

| Operation | nil channel | Open & not full / not empty | Empty (recv) | Full (send) | Closed |
|-----------|-------------|------------------------------|--------------|-------------|--------|
| **Read** | Block forever (→ deadlock) | value, true | Block | — | default value, false |
| **Write** | Block forever (→ deadlock) | OK | — | Block | **panic** |
| **Close** | **panic** | OK | OK | OK | **panic** (double close) |

Four ways to crash: close(nil), close(closed), send to closed, double-close. All solved by the **ownership pattern** below.

### Buffered channels
- Capacity-N in-memory FIFO queue.
- `make(chan T)` and `make(chan T, 0)` are both unbuffered.
- Send blocks when full; receive blocks when empty.
- ⚠️ Easily a **premature optimization** and can **hide deadlocks**.
- Special case: if buffer is empty and a receiver is already waiting, the sender hands the value directly to the receiver (skips the buffer).

### Channel ownership (the key pattern)
Define ownership as: the goroutine that **instantiates, writes to, and closes** the channel.

**Owner does:**
1. Instantiate the channel.
2. Perform writes (or transfer ownership).
3. Close the channel.
4. Expose it externally as receive-only (`<-chan T`).

**Consumer does:**
1. Check `ok` to know when the channel is closed.
2. Handle the fact that reads can block — decide on timeout, abandon, or block-forever semantics.

**Benefits of this pattern:**
- No writing to nil → no deadlock.
- No closing nil → no panic.
- No writing to closed → no panic.
- No double-close → no panic.
- Compile-time type-check prevents misuse.

```go
chanOwner := func() <-chan int {
    out := make(chan int, 5)
    go func() {
        defer close(out)
        for i := 0; i <= 5; i++ { out <- i }
    }()
    return out
}
```

---

## 4. The `select` Statement

The glue that composes channels together.

```go
select {
case v := <-c1:           // recv
case c2 <- x:             // send
case <-time.After(d):     // timeout
default:                  // non-blocking fallback
}
```

### Behavior
- Cases are evaluated **simultaneously**, not sequentially.
- If multiple cases are ready, Go picks one **pseudo-randomly** (roughly 50/50).
- If no case is ready and there's no `default` → the whole `select` blocks.
- If no case is ready and `default` exists → `default` runs immediately.

### Common patterns
- **Timeout:** `case <-time.After(1 * time.Second):` — `time.After` returns a channel that fires after the duration.
- **Non-blocking try:** add a `default:` to skip when nothing's ready.
- **For-select with `default`:** goroutine makes progress while polling for a done signal.
- **Block forever:** `select {}` (empty select, useful in main goroutine to keep program alive).
- **Disable a case dynamically:** set its channel to `nil` — a nil channel blocks forever, so `select` effectively skips it.

---

## 5. GOMAXPROCS

- Controls the number of OS threads hosting Go's work queues.
- **Pre Go 1.5:** defaulted to 1; people wrote `runtime.GOMAXPROCS(runtime.NumCPU())`.
- **Post Go 1.5:** automatically set to the number of logical CPUs. Usually don't touch it.

---

## Quick decision guide (interview-friendly)

| Situation | Use |
|-----------|-----|
| Wait for goroutines, don't care about results | `WaitGroup` |
| Need to collect results from goroutines | Channels + `select` |
| Guard shared mutable state | `Mutex` / `RWMutex` |
| Pass ownership / communicate between goroutines | Channels |
| Lazy one-time init | `Once` |
| Goroutine waits for an event | `Cond` (or a channel) |
| Reuse expensive objects | `Pool` |
| Timeout an operation | `select` + `time.After` |
| Disable a `select` case | Set its channel to `nil` |

## Memorize the four channel panics
1. `close(nil)` — panic
2. `close(closed)` — panic
3. Send to closed channel — panic
4. (Double close, a flavor of #2)

Plus the two deadlocks:
- Read from nil → blocks forever
- Send to nil → blocks forever

The **channel ownership pattern** eliminates all of these by construction.