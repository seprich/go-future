# go-future
[![Tests](https://github.com/seprich/go-future/actions/workflows/tests.yml/badge.svg?branch=main)](https://github.com/seprich/go-future/actions/workflows/tests.yml)
[![Latest Version](https://img.shields.io/github/v/tag/seprich/go-future?label=version&sort=semver)](https://github.com/seprich/go-future/tags)
[![gocard](https://goreportcard.com/badge/github.com/seprich/go-future)](https://goreportcard.com/report/github.com/seprich/go-future)
## Introduction
Lightweight Future abstraction over goroutines and return value channels.
The aim of this library is to introduce a simple model for running goroutines, and collecting results from them.

* **Features**
  * Execute any result returning function as a goroutine and return immediately a Future, which will receive the result.
  * Any panics that are not recovered on your worker function will be captured into the Future and re-released ("throws" panic) when you invoke the `.Await()` method of the future.
  * As goroutines cannot be forcefully cancelled, context.Context is not directly part of Future type, however the usage of Context is trivial to include seamlessly. See examples below.
* **Not supported**
  * Higher level functions such as map, flatMap, etc. In golang functional programming is an anti-pattern and this library does not aim to bring any fp concepts to go.

## Install
```shell
go get github.com/seprich/go-future
```
## Examples
```
import "github.com/seprich/go-future/async"
```

### Basic use
```go
package main

import (
	"fmt"
	"github.com/seprich/go-future/async"
)

func main() {
	f := async.NewFuture(func() (string, error) {
		return "Hello From GoRoutine !", nil
	})
	res, err := f.Await()
	if err != nil {
		fmt.Printf("Error: %v", err)
	}
	fmt.Printf("Result: %s", res)
}
```

### With Context
```go
package main

import (
	"fmt"
	"context"
	"github.com/seprich/go-future/async"
	"time"
)

func cancelableAsyncJob(ctx context.Context, name string) (string, error) {
	fmt.Println("Hello from " + name)
	for range 100 {
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("job canceled")
		default:
			time.Sleep(10 * time.Millisecond)	
        }
	}
	return "Done", nil
}

func main() {
	ctx, cancelFn := context.WithCancel(context.Background())

	fut := async.NewFuture2(cancelableAsyncJob, ctx, "Hi!")
	time.Sleep(10 * time.Millisecond)
	cancelFn()
	_, err := fut.Await()
	if err == nil {
		panic("error expected")
	}
}
```
