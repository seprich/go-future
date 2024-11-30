package async

import (
	"sync"
)

// Future - Light abstraction around executing and returning the result of a goroutine.
//
//	For Future without return values use a self-defined `type Nothing struct {}` => `Future[Nothing]`
type Future[T any] struct {
	mu sync.Mutex

	value    T
	err      error
	panicErr any

	done   chan struct{}
	isDone bool
}

/* CONSTRUCTORS */

// NewFuture - Execute result returning function as goroutine and return future for accessing the result.
func NewFuture[T any](fn func() (T, error)) *Future[T] {
	future := Future[T]{done: make(chan struct{})}
	go runAsync(&future, fn)
	return &future
}

// NewFuture1 - Variation for function taking 1 argument
func NewFuture1[T any, A any](fn func(arg1 A) (T, error), arg1 A) *Future[T] {
	future := Future[T]{done: make(chan struct{})}
	go runAsync1(&future, fn, arg1)
	return &future
}

// NewFuture2 - Variation for function taking 2 arguments
func NewFuture2[T any, A any, B any](fn func(arg1 A, arg2 B) (T, error), arg1 A, arg2 B) *Future[T] {
	future := Future[T]{done: make(chan struct{})}
	go runAsync2(&future, fn, arg1, arg2)
	return &future
}

// NewFuture3 - Variation for function taking 3 arguments
func NewFuture3[T any, A any, B any, C any](fn func(arg1 A, arg2 B, arg3 C) (T, error), arg1 A, arg2 B, arg3 C) *Future[T] {
	future := Future[T]{done: make(chan struct{})}
	go runAsync3(&future, fn, arg1, arg2, arg3)
	return &future
}

// NewFuture4 - Variation for function taking 4 arguments
func NewFuture4[T any, A any, B any, C any, D any](fn func(arg1 A, arg2 B, arg3 C, arg4 D) (T, error), arg1 A, arg2 B, arg3 C, arg4 D) *Future[T] {
	future := Future[T]{done: make(chan struct{})}
	go runAsync4(&future, fn, arg1, arg2, arg3, arg4)
	return &future
}

// NewFuture5 - Variation for function taking 5 arguments
func NewFuture5[T any, A any, B any, C any, D any, E any](fn func(arg1 A, arg2 B, arg3 C, arg4 D, arg5 E) (T, error), arg1 A, arg2 B, arg3 C, arg4 D, arg5 E) *Future[T] {
	future := Future[T]{done: make(chan struct{})}
	go runAsync5(&future, fn, arg1, arg2, arg3, arg4, arg5)
	return &future
}

// NewFuture6 - Variation for function taking 7 arguments
func NewFuture6[T any, A any, B any, C any, D any, E any, F any](fn func(arg1 A, arg2 B, arg3 C, arg4 D, arg5 E, arg6 F) (T, error), arg1 A, arg2 B, arg3 C, arg4 D, arg5 E, arg6 F) *Future[T] {
	future := Future[T]{done: make(chan struct{})}
	go runAsync6(&future, fn, arg1, arg2, arg3, arg4, arg5, arg6)
	return &future
}

// NewFutureCb - Callback version
func NewFutureCb[T any](fn func(resolve func(v T), reject func(e error))) *Future[T] {
	future := Future[T]{done: make(chan struct{})}
	go runAsyncCb(&future, fn)
	return &future
}

// NewFutureRace - Convenient alternative to using select
func NewFutureRace[T any](futures ...*Future[T]) *Future[T] {
	future := Future[T]{done: make(chan struct{})}
	for _, fut := range futures {
		go func(this *Future[T], source *Future[T]) {
			defer panicPacker(this)
			this.setResult(source.getResult())
		}(&future, fut)
	}
	return &future
}

/* INSPECTORS */

// HasResult - Check without blocking whether this future has resolved
func (f *Future[T]) HasResult() bool {
	return f.hasResult()
}

/* GETTERS */

// Await - Await for the Future to resolve and return the result
//
//	If the goroutine function panicked, this function will panic with identical content
func (f *Future[T]) Await() (T, error) { return f.getResult() }

// AwaitForDone - Await for execution done.
//
// Useful when goroutine does not return any meaningful values
func (f *Future[T]) AwaitForDone() error {
	_, err := f.getResult()
	return err
}

/* Private helpers - most importantly for tight control over mutex scoping */

func (f *Future[T]) setResult(r T, e error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if !f.isDone {
		f.value, f.err = r, e
		f.isDone = true
		close(f.done)
	}
}

func (f *Future[T]) setError(e error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if !f.isDone {
		f.err = e
		f.isDone = true
		close(f.done)
	}
}

func (f *Future[T]) setPanic(p any) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if !f.isDone {
		f.panicErr = p
		f.isDone = true
		close(f.done)
	}
}

func (f *Future[T]) hasResult() bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.isDone
}

func (f *Future[T]) getResult() (T, error) {
	<-f.done

	f.mu.Lock()
	val, err, pan := f.value, f.err, f.panicErr
	f.mu.Unlock()

	if pan != nil {
		panic(pan)
	}
	return val, err
}

func panicPacker[T any](f *Future[T]) {
	if r := recover(); r != nil {
		f.setPanic(r)
	}
}

func runAsync[T any](fut *Future[T], fn func() (T, error)) {
	defer panicPacker(fut)
	fut.setResult(fn())
}

func runAsync1[T any, A any](fut *Future[T], fn func(arg1 A) (T, error), arg1 A) {
	defer panicPacker(fut)
	fut.setResult(fn(arg1))
}

func runAsync2[T any, A any, B any](fut *Future[T], fn func(arg1 A, arg2 B) (T, error), arg1 A, arg2 B) {
	defer panicPacker(fut)
	fut.setResult(fn(arg1, arg2))
}

func runAsync3[T any, A any, B any, C any](fut *Future[T], fn func(arg1 A, arg2 B, arg3 C) (T, error), arg1 A, arg2 B, arg3 C) {
	defer panicPacker(fut)
	fut.setResult(fn(arg1, arg2, arg3))
}

func runAsync4[T any, A any, B any, C any, D any](fut *Future[T], fn func(arg1 A, arg2 B, arg3 C, arg4 D) (T, error), arg1 A, arg2 B, arg3 C, arg4 D) {
	defer panicPacker(fut)
	fut.setResult(fn(arg1, arg2, arg3, arg4))
}

func runAsync5[T any, A any, B any, C any, D any, E any](fut *Future[T], fn func(arg1 A, arg2 B, arg3 C, arg4 D, arg5 E) (T, error), arg1 A, arg2 B, arg3 C, arg4 D, arg5 E) {
	defer panicPacker(fut)
	fut.setResult(fn(arg1, arg2, arg3, arg4, arg5))
}

func runAsync6[T any, A any, B any, C any, D any, E any, F any](fut *Future[T], fn func(arg1 A, arg2 B, arg3 C, arg4 D, arg5 E, arg6 F) (T, error), arg1 A, arg2 B, arg3 C, arg4 D, arg5 E, arg6 F) {
	defer panicPacker(fut)
	fut.setResult(fn(arg1, arg2, arg3, arg4, arg5, arg6))
}

func runAsyncCb[T any](fut *Future[T], fn func(resolve func(v T), reject func(e error))) {
	defer panicPacker(fut)
	fn(func(v T) { fut.setResult(v, nil) }, func(e error) { fut.setError(e) })
}
