# pipeline-logger

A concurrent log processing pipeline built in Go, demonstrating goroutines and channels.

## What it does

Reads log lines and passes them through a 3-stage pipeline:

1. **Filter** — drops any line containing `DEBUG`
2. **Transform** — uppercases every line
3. **Output** — prints each line with a timestamp

Each stage runs in its own goroutine and communicates via channels. Channel ownership is explicit: each stage creates, writes to, and closes its own output channel.

## Run

**Pipe from a file:**
```bash
cat input.txt | go run main.go
```

**Send a string directly from the terminal:**
```bash
echo -e "INFO user logged in\nDEBUG checking cache\nERROR database failed" | go run main.go
```

**Expected output:**
```
INFO USER LOGGED IN
ERROR DATABASE FAILED
```

## Test

**Run tests:**
```bash
go test ./...
```

**Run tests with the race detector:**
```bash
go test -race ./...
```

## Concepts demonstrated

- Goroutine lifecycle and the `go` keyword
- Unbuffered channels for stage-to-stage communication
- Channel ownership: creator writes and closes
- `sync.WaitGroup` as a join point
- `io.Writer` injection for testability