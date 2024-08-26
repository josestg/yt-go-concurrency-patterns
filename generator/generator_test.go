package main

import (
	"context"
	"testing"
	"time"
)

func BenchmarkNewFibonacciStreamV3(b *testing.B) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	stream := NewFibonacciStreamV3(ctx)
	for i := 0; i < b.N; i++ {
		_, ok := <-stream
		if !ok {
			break
		}
	}
}

func BenchmarkNewFibonacciStreamV4(b *testing.B) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	stream := NewFibonacciStreamV4(ctx, 10)
	for i := 0; i < b.N; i++ {
		_, ok := <-stream
		if !ok {
			break
		}
	}
}
