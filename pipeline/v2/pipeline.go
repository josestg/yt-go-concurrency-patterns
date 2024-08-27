package main

import (
	"context"
	"fmt"
	"time"
)

type (
	// Operator does some operation on the input stream before passing it to the output stream.
	Operator[T any] func(ctx context.Context, in <-chan T) (out <-chan T)

	// TerminalOperator consumes the data from the input stream one at a time.
	TerminalOperator[T any] func(ctx context.Context, in <-chan T)
)

var _ Transform[int] = Triple
var _ Transform[int] = Successor

func Triple(e int) int { return e * 3 }

func Successor(e int) int { return e + 1 }

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	piped := Pipeline(
		ctx,
		StreamOf(ctx, 1, 2, 3, 4, 5),
		Delay[int](50*time.Millisecond),
		Filter(IsOdd),
		Map(Triple),
		Map(Successor),
	)
	ForEach(ctx, piped)
	// Output: (close order may vary)
	// 4
	// 10
	// Filter: closed
	// StreamOf: closed
	// Map: closed
	// Map: closed
	time.Sleep(time.Second)
}

// Pipeline creates a pipeline of operators to process the input stream.
func Pipeline[T any](ctx context.Context, source <-chan T, operators ...Operator[T]) <-chan T {
	stream := source
	for _, operator := range operators {
		stream = operator(ctx, stream)
	}
	return stream
}

// Ensure that ForEach follows the TerminalOperator signature.
var _ TerminalOperator[any] = ForEach

func ForEach[T any](ctx context.Context, in <-chan T) {
	for v := range in {
		println(v)
	}
}

// StreamOf creates a stream of items of type T.
func StreamOf[T any](ctx context.Context, seq ...T) <-chan T {
	stream := make(chan T)
	go func() {
		defer func() {
			close(stream)
			fmt.Println("StreamOf: closed")
		}()
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

// Transform transforms the input stream using the transform function.
type Transform[T any] func(T) T

// Map applies the transform function to each item in the input stream.
func Map[T any](transform Transform[T]) Operator[T] {
	return func(ctx context.Context, in <-chan T) <-chan T {
		out := make(chan T)
		go func() {
			defer func() {
				close(out)
				fmt.Println("Map: closed")
			}()
			for v := range in {
				select {
				case <-ctx.Done():
					return
				case out <- transform(v):
				}
			}
		}()
		return out
	}
}

// Predicate is a function that checks if a condition is satisfied.
type Predicate[T any] func(T) bool

// Filter filters out items from the input stream that do not satisfy the predicate p.
func Filter[T any](p Predicate[T]) Operator[T] {
	return func(ctx context.Context, in <-chan T) <-chan T {
		out := make(chan T)
		go func() {
			defer func() {
				close(out)
				fmt.Println("Filter: closed")
			}()
			for v := range in {
				if p(v) {
					select {
					case <-ctx.Done():
						return
					case out <- v:
					}
				}
			}
		}()
		return out
	}
}

var _ Predicate[int] = IsOdd
var _ Predicate[int] = IsEven

// IsOdd is a predicate that checks if a number is odd.
func IsOdd(n int) bool { return n%2 == 1 }

// IsEven is a predicate that checks if a number is even.
func IsEven(n int) bool { return n%2 == 0 }

// Delay delays the emission of items from the input stream by d duration.
func Delay[T any](d time.Duration) Operator[T] {
	return func(ctx context.Context, in <-chan T) (out <-chan T) {
		stream := make(chan T)
		go func() {
			defer close(stream)
			for {
				select {
				case <-ctx.Done():
					return
				case <-time.After(d):
					select {
					case <-ctx.Done(): // recheck deadline after delay.
						return
					case stream <- <-in:
					}
				}
			}
		}()
		return stream
	}
}
