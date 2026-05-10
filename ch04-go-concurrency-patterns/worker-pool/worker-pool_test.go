package worker_pool

import (
	"context"
	"errors"
	"reflect"
	"sort"
	"testing"
	"time"
)

func TestRunPoolProcessesAllJobs(t *testing.T) {
	ctx := context.Background()

	jobs := make(chan int)
	results := RunPool(ctx, jobs, 3, func(ctx context.Context, job int) (int, error) {
		return job * 2, nil
	})

	go func() {
		defer close(jobs)

		for i := 1; i <= 5; i++ {
			jobs <- i
		}
	}()

	var got []int

	for result := range results {
		if result.Error != nil {
			t.Fatalf("expected no error, got %v", result.Error)
		}

		got = append(got, result.Value)
	}

	sort.Ints(got)

	want := []int{2, 4, 6, 8, 10}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestRunPoolReturnsErrors(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("failed job")

	jobs := make(chan int)
	results := RunPool(ctx, jobs, 2, func(ctx context.Context, job int) (int, error) {
		if job == 2 {
			return 0, expectedErr
		}

		return job, nil
	})

	go func() {
		defer close(jobs)

		jobs <- 1
		jobs <- 2
		jobs <- 3
	}()

	var values []int
	var errs []error

	for result := range results {
		if result.Error != nil {
			errs = append(errs, result.Error)
			continue
		}

		values = append(values, result.Value)
	}

	sort.Ints(values)

	if !reflect.DeepEqual(values, []int{1, 3}) {
		t.Fatalf("expected values %v, got %v", []int{1, 3}, values)
	}

	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}

	if !errors.Is(errs[0], expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, errs[0])
	}
}

func TestRunPoolClosesResultStreamWhenJobsClose(t *testing.T) {
	ctx := context.Background()

	jobs := make(chan int)
	results := RunPool(ctx, jobs, 2, func(ctx context.Context, job int) (int, error) {
		return job, nil
	})

	close(jobs)

	select {
	case _, ok := <-results:
		if ok {
			t.Fatal("expected result stream to be closed")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("result stream was not closed")
	}
}

func TestRunPoolStopsWhenContextIsCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	jobs := make(chan int)

	results := RunPool(ctx, jobs, 2, func(ctx context.Context, job int) (int, error) {
		return job, nil
	})

	cancel()

	select {
	case _, ok := <-results:
		if ok {
			t.Fatal("expected result stream to be closed")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("result stream was not closed after context cancellation")
	}
}

func TestRunPoolDoesNotBlockWhenConsumerStopsAndContextIsCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	jobs := make(chan int)

	results := RunPool(ctx, jobs, 1, func(ctx context.Context, job int) (int, error) {
		return job, nil
	})

	jobs <- 1

	cancel()

	select {
	case <-results:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("worker did not stop after context cancellation")
	}
}

func TestRunPoolRunsJobsConcurrently(t *testing.T) {
	ctx := context.Background()

	jobs := make(chan int)

	start := time.Now()

	results := RunPool(ctx, jobs, 3, func(ctx context.Context, job int) (int, error) {
		time.Sleep(100 * time.Millisecond)
		return job, nil
	})

	go func() {
		defer close(jobs)

		for i := 0; i < 3; i++ {
			jobs <- i
		}
	}()

	var count int

	for range results {
		count++
	}

	if count != 3 {
		t.Fatalf("expected 3 results, got %d", count)
	}

	elapsed := time.Since(start)

	if elapsed >= 250*time.Millisecond {
		t.Fatalf("expected jobs to run concurrently, took %v", elapsed)
	}
}
