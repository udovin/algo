package futures

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
)

type Future[T any] interface {
	Get(ctx context.Context) (T, error)
}

type ResultSetter[T any] func(T, error)

func New[T any]() (Future[T], ResultSetter[T]) {
	done := make(chan struct{})
	f := future[T]{done: done}
	once := sync.Once{}
	setResult := func(value T, err error) {
		once.Do(func() {
			f.value, f.err = value, err
			close(done)
		})
	}
	return &f, setResult
}

func Call[T any](fn func() (T, error)) Future[T] {
	f, setResult := New[T]()
	go func() {
		panicking := true
		defer func() {
			if r := recover(); panicking {
				var empty T
				setResult(empty, PanicError{
					Value: r,
					Stack: debug.Stack(),
				})
			}
		}()
		setResult(fn())
		panicking = false
	}()
	return f
}

func CallAfter[T any, V any](future Future[T], fn func(Future[T]) (V, error)) Future[V] {
	wrapFn := func() (V, error) {
		return fn(future)
	}
	return Call(wrapFn)
}

type future[T any] struct {
	done  <-chan struct{}
	value T
	err   error
}

func (f *future[T]) Get(ctx context.Context) (T, error) {
	select {
	case <-f.done:
		return f.value, f.err
	default:
	}
	select {
	case <-f.done:
		return f.value, f.err
	case <-ctx.Done():
		return f.value, ctx.Err()
	}
}

type PanicError struct {
	Value any
	Stack []byte
}

func (p PanicError) Error() string {
	return fmt.Sprintf("panic: %v", p.Value)
}
