# Chapter 1
## Common issues when working with concurrent code
### Race conditions
A race condition occurs when two or more operations mmust execute in the correct order
but the program has not been written so that this order is guaranteed to be maintained.
Most of the time this shows up in a *data race*, where one concurrent operation attempts to read a variable while at some undetermined time another concurrent operation is attempting to write to the same variable.

This issues mostly happen because **the developer is thinking about the problem sequentially**

### Atomicity
Something is considered atomic, or to have the property of atomicity, when it is indivisible or uninterruptible in the context that it is operating.
It is very important to consider the **context** that an atomic process is being executed on. This fact can work both for and against us.

Combining atomic operations does not mean a bigger atomic operation, example:

`i++`

This can look atomic, but in fact this three things are happening:

- Retrieve the value of i
- Increment the value of i
- Store the value of i

If this operation happens in the context of a program with no concurrent processes, this code is atomic within that context.

If the context is a goroutine that does not expose i to other goroutines, then this code is atomic.

### Memory Access Syncronization
We can use locks to make code have exclusive access to the data, this can help us solve data races but it isn't the most performant way and has some troubles like expecting all developers to do the same everywhere.

### Deadlocks
A deadlock program is one in which all concurrent processes are waiting on one another. In this state, the program cannot recover without outside invervention.

There are a few conditions that must be present for deadlocks to arise, they are known as *Coffman Conditions*

- *Mutual Exclusion*: A concurrent process holds exclusive rights to a resource at any one time.
- *Wait For Condition*: A concurrent process must simultaneously hold a resource and be waiting for an additional resource.
- *No Preemption*: A resource held by a concurrent process can only be released by that process, so it fulfills this condition.
- *Circular Wait*: A concurrent process (P1) must be waiting on a chain of other concurrent processes (P2) which are in turn waiting on it (P1), so it fulfills this final condition too.

### Livelock
Livelocks are like deadlocks, but instead of waiting for another process, they both answer to each other, and that's why it never makes progress.
Real world analogy: two people in a corridor walking toward each other. Both step aside to let the other pass, but they both step to the same side. Then both step to the other side. Then back again. Neither is frozen. Both are moving. Nobody gets through.
The process looks alive because both processes are doing things, but they are not doing them correctly.
They are harder to detect because on the surface they can look correct as if you look at the CPU utilization to determine if it was doing anything, you might think it was.
**Livelocks are a subset of a larger set of problems called *starvation***.

### Starvation
Starvation implies that there are one or more greedy concurrent process that are unfairly preventing one or more concurrent processes from accomplishing work as efficiently as possible, or maybe at all.
One example could be a greedy worker holding onto a shared lock for the entirety of its work loop, whereas a polite worker attempts to only lock when it needs to. This makes the greedy lock do more operations than the polite one, which is not efficient.

A good way of identifying starvation is to record and sample metrics. One of the ways we can detect and solve starvation is by logging when work is accomplished, and then determining if the rate of work is as high as expected through all workers.

Any resource that must be shared is a candidate for starvation.

#### Tip
When using memory access synchronization, we have to find a balance between coarse-grained synchronization for performance, and fine-grained synchronization for fairness. To performance tune our application, to start with, it's highly recommended to constrain memory access synchronization only
to critical sections. If the synchronization becomes a performance problem, we can always broaden the scope. It's much harder to go the other way around.

### Determining Concurrency Safety
This is a hard problem, it already is in greenfield projects, but it can be even harder in projects with more developers that have worked on the project for years.

Example:

```
// CalculatePi calculates digits of Pi between the begin and end place.
func CalculatePi(begin, end int64, pi *Pi)
```

Calculating pi is best done concurrently, but this example raises a lot of questions:

- How do I do so with this function?
- Am I responsible for instantiating multiple concurrent invocations?
- Should I synchronize access to the address of Pi? Or is it handled inside the function?

We can make this simpler to understand with more comments focusing on clarity:
```
// CalculatePi calculates digits of Pi between the begin and end place.
//
// Internally, CalculatePi will create FLOOR((end-begin)/2) concurrent
// processes which recursively call CalculatePi. Synchronization of
//writes to pi are handled internally by the Pi struct.
func CalculatePi(begin, end int64, pi *Pi)
```

We now understand that we can call the function plainly and not worry about concurrency or synchronization.

We should always cover the following aspects when writting a function that uses concurrency

- Who is responsible for the concurrency?
- How is the problem space mapped onto concurrency primitives?
- Who is responsible for the synchronization?

Try to always err on the side of verbose comments and cover this three aspects. Ambiguity in this function suggests that we have modeled it wrong, maybe we should
take a functional approach and ensure our function has no side effects:
```
func CalculatePi(begin, end int64) []uint
```

This signature removes any questions about synchronization, but still leaves the question of whether concurrency is used. We can odify it again to throw out another signal as to what is happening:
```
func CalculatePi(begin, end int64) <-chan uint
```

The return implies that the function is doing work asynchronously and handing us a stream