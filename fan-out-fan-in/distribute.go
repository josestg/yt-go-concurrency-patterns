package main

import (
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
func NewFibonacciStream(in <-chan int64) <-chan fib {
	out := make(chan fib)
	go func() {
		defer close(out)
		for v := range in {
			out <- fib{
				n:      v,
				result: SlowFibonacci(v),
			}
		}
	}()
	return out
}

func StreamOf[T any](seq ...T) <-chan T {
	stream := make(chan T)
	go func() {
		defer close(stream)
		for _, v := range seq {
			stream <- v
		}
	}()
	return stream
}

func main() {
	stream := StreamOf[int64](40, 41, 42, 43, 44)
	started := time.Now()

	for v := range Distribute(stream, NewFibonacciStream, 5) {
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

func Distribute(
	source <-chan int64,
	task func(source <-chan int64) <-chan fib,
	replicas int,
) <-chan fib {
	consumers := make([]<-chan fib, replicas)
	for i := 0; i < replicas; i++ {
		consumers[i] = task(source)
	}
	return Merge(consumers...)
}

func Merge(in ...<-chan fib) <-chan fib {
	var wg sync.WaitGroup

	out := make(chan fib)
	worker := func(ch <-chan fib) {
		defer wg.Done()
		for v := range ch {
			out <- v
		}
	}

	wg.Add(len(in))
	for _, stream := range in {
		go worker(stream)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
