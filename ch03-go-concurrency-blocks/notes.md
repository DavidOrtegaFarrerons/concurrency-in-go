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