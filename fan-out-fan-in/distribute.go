package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// SlowFibonacci calculates the nth Fibonacci number.
func SlowFibonacci(n int64) int64 {
	if n <= 1 {
		return n
	}
	return SlowFibonacci(n-1) + SlowFibonacci(n-2)
}

// fib contains the input number n and the result of the Fibonacci calculation.
type fib struct{ n, result int64 }

// NewFibonacciStream creates a stream of Fibonacci numbers.
func NewFibonacciStream(ctx context.Context, in <-chan int64) <-chan fib {
	out := make(chan fib)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case n, ok := <-in:
				if !ok {
					return
				}
				out <- fib{n: n, result: SlowFibonacci(n)}
			}
		}
	}()
	return out
}

// StreamOf creates a stream of items of type T.
func StreamOf[T any](ctx context.Context, seq ...T) <-chan T {
	stream := make(chan T)
	go func() {
		defer close(stream)
		for _, item := range seq {
			select {
			case <-ctx.Done():
				return
			case stream <- item:
			}

		}
	}()
	return stream
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	stream := StreamOf[int64](ctx, 40, 41, 42, 43, 44)
	started := time.Now()

	for v := range Distribute(ctx, stream, NewFibonacciStream, 2) {
		fmt.Printf("%+v\n", v)
	}
	fmt.Printf("Elapsed: %v\n", time.Since(started))
	// Output:
	// {n:40 result:102334155}
	// {n:41 result:165580141}
	// {n:42 result:267914296}
	// {n:43 result:433494437}
	// {n:44 result:701408733}
	// Elapsed: 2.401358083s
}

// Distribute distributes the input stream to multiple workers.
func Distribute[InpStream ~<-chan T, OutStream ~<-chan U, T, U any](
	ctx context.Context,
	s InpStream,
	worker func(ctx context.Context, s InpStream) OutStream,
	replicas int,
) OutStream {
	consumers := make([]OutStream, replicas)
	for i := 0; i < replicas; i++ {
		consumers[i] = worker(ctx, s)
	}
	return Merge(ctx, consumers...)
}

// Merge merges multiple streams into a single stream.
func Merge[Stream ~<-chan T, T any](ctx context.Context, sources ...Stream) Stream {
	var wg sync.WaitGroup
	out := make(chan T)

	worker := func(ch Stream) {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-ch:
				if !ok {
					return
				}
				out <- v
			}
		}
	}

	wg.Add(len(sources))
	for _, stream := range sources {
		go worker(stream)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
