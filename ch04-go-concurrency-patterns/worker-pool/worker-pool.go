package worker_pool

import (
	"context"
	"sync"
)

type Result[R any] struct {
	Value R
	Error error
}

// Using orDone creates 10 extra goroutines to proxy the same jobs channel, so we don't use it
func orDone[T any](ctx context.Context, jobs <-chan T) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-jobs:
				if !ok {
					return
				}

				select {
				case out <- v:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}

func RunPool[T, R any](
	ctx context.Context,
	jobs <-chan T,
	n int,
	fn func(context.Context, T) (R, error),
) <-chan Result[R] {
	resultStream := make(chan Result[R])
	wg := sync.WaitGroup{}

	workerJob := func(ctx context.Context, jobs <-chan T) {
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				return
			case job, ok := <-jobs:
				if !ok {
					return
				}

				r, err := fn(ctx, job)
				result := Result[R]{Value: r, Error: err}

				select {
				case <-ctx.Done():
					return
				case resultStream <- result:
				}

			}
		}
	}

	for range n {
		wg.Add(1)
		go workerJob(ctx, jobs)
	}

	go func() {
		defer close(resultStream)
		wg.Wait()
	}()

	return resultStream
}
