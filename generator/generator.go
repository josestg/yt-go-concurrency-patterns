package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	{
		fmt.Println("v0: Fibonacci Generator")
		fib := NewFibonacciStreamV0()
		for i := 0; i < 10; i++ {
			fmt.Printf("v0: fib(%d)=%d\n", i, fib())
		}
	}
	{
		fmt.Println("v1: Fibonacci Stream")
		stream := NewFibonacciStreamV1()
		for i := 0; i < 10; i++ {
			fmt.Printf("v1: fib(%d)=%d\n", i, <-stream)
		}
	}

	{
		fmt.Println("v2: Fibonacci Stream with explicit cancel")
		stream, cancel := NewFibonacciStreamV2()
		for i := 0; i < 10; i++ {
			if i == 5 {
				cancel()
			}
			v, ok := <-stream
			if !ok {
				fmt.Println("v2: stream closed")
				break
			}
			fmt.Printf("v2: fib(%d)=%d\n", i, v)
		}
	}

	{
		fmt.Println("v3: Fibonacci Stream with context-based cancel")
		ctx, cancel := context.WithCancel(context.Background())
		stream := NewFibonacciStreamV3(ctx)
		for i := 0; i < 10; i++ {
			if i == 5 {
				cancel()
			}
			v, ok := <-stream
			if !ok {
				fmt.Println("v3: stream closed")
				break
			}
			fmt.Printf("v3: fib(%d)=%d\n", i, v)
		}
	}

	{
		fmt.Println("v3 + timeout: Fibonacci Stream with context-based cancel")
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Microsecond)
		defer cancel()
		stream := NewFibonacciStreamV3(ctx)
		for i := 0; ; i++ {
			v, ok := <-stream
			if !ok {
				fmt.Println("v3 + timeout: stream closed")
				break
			}
			fmt.Printf("v3 + timeout: fib(%d)=%d\n", i, v)
		}
	}

	{
		fmt.Println("v4: Fibonacci Stream with context-based cancel and buffer")
		ctx, cancel := context.WithCancel(context.Background())
		stream := NewFibonacciStreamV4(ctx, 5)
		for i := 0; i < 20; i++ {
			if i == 10 {
				cancel()
			}
			v, ok := <-stream
			if !ok {
				fmt.Println("v4: stream closed")
				break
			}
			fmt.Printf("v4: fib(%d)=%d\n", i, v)
		}
	}
}

func NewFibonacciStreamV0() func() int {
	a, b := 0, 1
	return func() int {
		defer func() { a, b = b, a+b }()
		return a
	}
}

func NewFibonacciStreamV1() <-chan int {
	stream := make(chan int)
	go func() {
		a, b := 0, 1
		for {
			stream <- a
			a, b = b, a+b
		}
	}()
	return stream
}

func NewFibonacciStreamV2() (<-chan int, func()) {
	stream := make(chan int)
	quit := make(chan struct{})
	go func() {
		defer close(stream)
		a, b := 0, 1
		for {
			select {
			case <-quit:
				return
			case stream <- a:
				a, b = b, a+b
			}
		}
	}()

	cancel := func() { close(quit) }
	return stream, cancel
}

func NewFibonacciStreamV3(ctx context.Context) <-chan int {
	stream := make(chan int)
	go func() {
		defer close(stream)
		a, b := 0, 1
		for {
			select {
			case <-ctx.Done():
				return
			case stream <- a:
				a, b = b, a+b
			}
		}
	}()
	return stream
}

func NewFibonacciStreamV4(ctx context.Context, bufsize int) <-chan int {
	stream := make(chan int, bufsize)
	go func() {
		defer close(stream)
		a, b := 0, 1
		for {
			select {
			case <-ctx.Done():
				return
			case stream <- a:
				a, b = b, a+b
			}
		}
	}()
	return stream
}
