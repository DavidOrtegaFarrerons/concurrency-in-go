# Chapter 3
## Goroutines
Every program has at least 1 goroutine, the main goroutine, which is automatically created and started when the process begins.

A goroutine is a function that runs concurrently (not necessarily in parallel) alongside other code.
To start one, we can simply put the `go` keyword before a function

```
func man() {
    go sayHello()
}

func sayHello() {
    fmt.Println("hello")
}
```

it also works with anonymous functions:

```
go func() {
    fmt.Println("hello")
}()
```

Notice that we must invoke the anonymous function immediately to use the `go` keyword

Alternatively, we can assign the function to a variable and call the anonymous function like this:
```
sayHello := func() {
    fmt.Println("hello")
}
go sayHello()
```

Goroutines are not OS threads, and they are not exactly green threads (threads managed by the language's runtime).
They are a higher level of abstraction known as *coroutines*. Coroutines are concurrent subroutines (functions, closures, or methods in Go)
that are *nonpreemtive* (they cannot be interrupted). Coroutines have multiple points throughout which allow for suspension or reentry.

Goroutines don't define their own suspension or reentry points, Go's runtime obersves the runtime behavior of goroutines and automatically suspends them when they block and then resumes them when they become unblocked.
They are implicitly concurrent constructs, but concurrency is not a property of a coroutine: something must host several coroutines simultaneously and give each an opportunity to execute, otherwhise, they wouldn't be concurrent!
This does not imply that coroutines are implicitly parallel.
It is possible to have several coroutines executing sequentially to give the illusion of parallelism, and in fact this happens all the time in Go

Go's mechanism for hosting goroutines is an implementation of what's called an M:M scheduler.
It maps M green threads to N OS threads. Goroutines are then scheduled onto the green threads.
When we have more goroutines than green threads available, the scheduler handles the distribution of the goroutines across the available threads and ensures that when these goroutines become blocked, other goroutines can be run.

Go follows a model of concurrency called the *fork-join* model.
The word *fork* refers to the fact that at any point in the program, it can split off a *child* branch of execution to be run concurrently with its *parent*.
The world *join* refers to the fact that at some point in the future, these concurrent branches of execution will join back together.
Where the child rejoins the parent is called a *join point*.

Join pints are what guarantee our program's correctness and remove the race condition. To create a join point, we have to synchronize the main goroutine and the child goroutine.
For example:

```
var wg sync.WaitGroup
sayHello := func() {
    defer wg.Done()
    fmt.Println("hello")
}
wg.Add(1)
go sayHello()
wg.Wait() //This is the join point
```

If you run a closure in a goroutine, does the closure operate on a copy of these variables, or the original reference?
```
var wg sync.WaitGroup
salutation := hello
wg.Add(1)
go func() {
    defer wg.Done()
    salutation = "welcome"
}()
wg.Wait()
fmt.Println(salutation)
```

The outcome in this case is "welcome"

Goroutines execute within the same address space they were created in.

```
var wg sync.WaitGroup
for _, salutation := range []string{"hello", "greetings", "good day"} {
    wg.Add(1)
    go func() {
        defer wg.Done()
        fmt.Println(salutation)
    }()
}
wg.Wait()
```

The outcome in this case is "good day" three times.

This happens because there is a high probability that the loop will exit before the goroutines are begun.
This means the *salutation* variable falls out of scope. Instead of this memory being garbage collected, Go runtime is observant enough to know that a reference to the salutation variable is still being held, and therefore will transfer the memory to the heap so that the goroutines can continue to access it.
As the loop will usually exit before any goroutines begin running, the value of salutation will be the last of the for loop ("good day"), which is the one copied to the heap and therefore showing the "good day" three times.

The proper way to write this loop is to pass a copy of salutation into the closure so that by the time the goroutine is run, it will be operating on the data from its iteration of the loop:
```
var wg sync.WaitGroup
for _, salutation := range []string{"helo", "greetings", "good day"} {
    wg.Add(1)
    go func(salutation string) {
        defer wg.Done()
        fmt.Println(salutation)
    }(salutation)
}
wg.Wait()
```

As we pass the current iteration variable to the closure, a copy of the string struct is made, thereby ensuring that when the goroutine is run, we refer to the proper string.

Goroutines are given a few kilobytes, which is almost always enough, when it isn't, the run-time grows (and shrinks) the memory for storing the stack automatically.

### The sync Package
#### WaitGroup
WaitGroup is a great way to wait for a set of concurrent operations to complete when we either don't care about the result of the concurrent operation, or we have other means of collecting their results.
If none of those conditions are true, we should use channels and a select statement instead.

Example of using a WaitGroup to wait for goroutines to complete:
```
var wg sync.WaitGroup

wg.Add(1)
go func() {
    defer wg.Done()
    fmt.Println("1st goroutine sleeping...")
    time.Sleep(1)
}()

wg.Add(1)
go func() {
    defer wg.Done()
    fmt.Println("2st goroutine sleeping...")
    time.Sleep(2)
}()

wg.Wait()
fmt.Println("All goroutines complete.")
```

We use `wg.Wait()` with an argument of 1 to indicate that one goroutine is beginning.
We use `defer wg.Done()` to ensure that bfore we exit the goroutine's closure, we indicate to the WaitGroup that we've exited.
Finally, we call `wg.Wait()` to block the main goroutine untill all goroutines have indicated that they have exited.

We have to call `Add()` outside the goroutine itself, otherwise, we would introduce a race condition.
We should couple calls to `Add()` as close as possible to the goroutines they're helping to track. We can also use it to track a group of goroutines all at once.


#### Mutex and RWMutex (mutually exclusive)
Mutex stands for "mutual exclusion", and is a way to guard critical sections of your program. A critical section is an area of our program that requires exclusive access to a shared resource.
A Mutex provides concurrent-safe way to express exclusive access to these shared resources.

```
var count int
var lock sync.Mutex

increment := func() {
    lock.Lock()
    defer lock.Unlock()
    count++
    fmt.Println("Incrementing: %d\n", count)
}
```

Using `lock.Lock()` we request the exclusive use of the critical section, in this case the count variable, guarded by a Mutex, lock.
Using `lock.Unlock()` we indicate that we are done with the critical section lock is guarding. Using `defer`to call `Unlock`is a very common idiom when using Mutex to ensure the call always happens, even when panicking. Not doing so will probably cause your program to have a deadlock

Critical sections are so named because they reflect a bottleneck in our program. It is expensive to enter and exit a critical section, and generally people attempt to minimize the time spent in critical sections.
One strategy is to check if we need all the processes to read *and* write to this memory. If not, we can take advantage of `sync.RWMutex`.

`sync.RWMutex` is conceptually the same as a `Mutex`, it guards access to memory, however, `RWMutex` gives us a bit more control over the memory. We can request a lock for reading, in which case we will be granted access unless the lock is being held for writing.
Example:

It is usually advisable to use RWMutex instead of Mutex when it logically makes sense.

### Cond
```
c := sync.NewCond(&sync.Mutex{})
c.L.Lock()
for conditionTrue() == false {
    c.Wait()
}
c.L.Unlock()
```

- `sync.NewCond` takes any type that satisfies the sync.Locker interface.
- We call `c.L.Lock()`, this is necessary because the call to `Wait()` automatically calls `Unlock` on the `Locker` when entered
- When the `Wait()` exits, it locks the `Locker` lock

A call to `Wait` doesn't just block, it *suspends* the current goroutine, allowing other goroutines to run on the OS thread.

Upon entering `Wait`, `Unlock` is called on the Cond variable's `Locker`, and upon exiting `Wait`, `Lock` is called on the Cond variable's `Locker`. Even though it loookes like we are holding this lock the entier time while we wait for the condition to ocurr, that's not actually the case.

`Signal()` is a method from `Cond` that maintains a FIFO list of goroutines waiting to be signaled. Signal finds the goroutine that's been waiting the longest and notifies that.

`Broadcast()` sends a signal to all goroutines that are waiting.

#### Once

`once.Do()` will execute the function passed in exactly once. Always once.
