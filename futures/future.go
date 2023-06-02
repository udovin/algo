package futures

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
)

type Future[T any] interface {
	Get(ctx context.Context) (T, error)
	Done() <-chan struct{}
}

type ResultSetter[T any] func(value T, err error)

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

func NewDone[T any](value T, err error) Future[T] {
	return doneFuture[T]{value: value, err: err}
}

type PanicError struct {
	Value any
	Stack []byte
}

func (p PanicError) Error() string {
	return fmt.Sprintf("panic: %v", p.Value)
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
	var empty T
	select {
	case <-f.done:
		return f.value, f.err
	case <-ctx.Done():
		return empty, ctx.Err()
	}
}

func (f *future[T]) Done() <-chan struct{} {
	return f.done
}

type doneFuture[T any] struct {
	value T
	err   error
}

func (v doneFuture[T]) Get(ctx context.Context) (T, error) {
	return v.value, v.err
}

func (v doneFuture[T]) Done() <-chan struct{} {
	return chanDone
}

var chanDone = make(chan struct{})

func init() {
	close(chanDone)
}
