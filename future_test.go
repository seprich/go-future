package async

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"
)

func TestBasic(t *testing.T) {
	f := NewFuture(func() (string, error) {
		return "happy days", nil
	})
	r, err := f.Await()
	if err != nil {
		t.Errorf("Didn't expect errors, got %s", err)
	}
	if r != "happy days" {
		t.Errorf("Expected '%s', got '%s'", "happy days", r)
	}
}

func TestParallelExcutions(t *testing.T) {
	assertOk := assertNoError[int](t)
	f1 := NewFuture1(asyncAddOne, 1)
	f2 := NewFuture1(asyncAddOne, 2)
	f3 := NewFuture1(asyncAddOne, 3)
	r1 := assertOk(f1.Await())
	r2 := assertOk(f2.Await())
	r3 := assertOk(f3.Await())
	if r1 != 2 || r2 != 3 || r3 != 4 {
		t.Error("Incorrect results")
	}
}

func TestAccessResultMultipleTimes(t *testing.T) {
	fut := NewFuture1(asyncAddOne, 10)
	r1 := assertNoError[int](t)(fut.Await())
	r2 := assertNoError[int](t)(fut.Await())
	if r1 != 11 || r2 != 11 {
		t.Error("Incorrect results")
	}
}

func TestCancellableFuture(t *testing.T) {
	ctx, cancelFn := context.WithCancel(context.Background())
	fut := NewFuture2(asyncCancelable, ctx, "one")
	if fut.hasResult() {
		t.Error("Unexpectedly future claims to have result")
	}
	time.Sleep(30 * time.Millisecond)
	cancelFn()
	res, err := fut.Await()
	if err == nil || res == 0 || res == 1000 {
		t.Error("Unexpectedly cancel seems to have failed")
	}
}

func TestRaceAndCancel(t *testing.T) {
	ctx, cancelFn := context.WithCancel(context.Background())
	f1 := NewFuture2(asyncCancelable, ctx, "one")
	f2 := NewFuture2(asyncCancelable, ctx, "two")
	f3 := NewFuture2(asyncCancelable, ctx, "quick")
	rf := NewFutureRace(f1, f2, f3)
	res, err := rf.Await()
	cancelFn() // Result has been delivered cancel the slow jobs still running
	if err != nil || res != 2000 {
		t.Error("Unexpected end result")
	}
	e1 := f1.AwaitForDone()
	e2 := f2.AwaitForDone()
	if e1 == nil || e1.Error() != "cancel one" {
		t.Error("Unexpected result from f1")
	}
	if e2 == nil || e2.Error() != "cancel two" {
		t.Error("Unexpected result from f2")
	}
}

func Test3thArityFuture(t *testing.T) {
	asyncFn := func(a string, b int, c int) (string, error) {
		time.Sleep(5 * time.Millisecond)
		return a + strconv.Itoa(b+c), nil
	}
	fut := NewFuture3(asyncFn, "hello ", 9, 4)
	expected := "hello 13"
	res := assertNoError[string](t)(fut.Await())
	if res != expected {
		t.Errorf("Expected: %s, Got: %s", expected, res)
	}
}

func Test4thArityFuture(t *testing.T) {
	asyncFn := func(a string, b int, c int, d bool) (string, error) {
		time.Sleep(5 * time.Millisecond)
		return fmt.Sprintf("%s %d %t", a, b+c, d), nil
	}
	fut := NewFuture4(asyncFn, "hello", 4, 2, true)
	expected := "hello 6 true"
	res := assertNoError[string](t)(fut.Await())
	if res != expected {
		t.Errorf("Expected: %s, Got: %s", expected, res)
	}
}

func Test5thArityFuture(t *testing.T) {
	asyncFn := func(a string, b int, c int, d bool, e float32) (string, error) {
		time.Sleep(5 * time.Millisecond)
		return fmt.Sprintf("%s %d %t %f", a, b+c, d, e), nil
	}
	fut := NewFuture5(asyncFn, "hi", 45, 21, true, 3.5)
	expected := "hi 66 true 3.500000"
	res := assertNoError[string](t)(fut.Await())
	if res != expected {
		t.Errorf("Expected: %s, Got: %s", expected, res)
	}
}

func Test6thArityFuture(t *testing.T) {
	asyncFn := func(a string, b int, c int, d bool, e float32, f string) (string, error) {
		time.Sleep(5 * time.Millisecond)
		return fmt.Sprintf("%s %d %t %f %s", a, b+c, d, e, f), nil
	}
	fut := NewFuture6(asyncFn, "hi", 45, 21, true, 3.5, "end")
	expected := "hi 66 true 3.500000 end"
	res := assertNoError[string](t)(fut.Await())
	if res != expected {
		t.Errorf("Expected: %s, Got: %s", expected, res)
	}
}

func assertNoError[T any](t *testing.T) func(res T, err error) T {
	return func(res T, err error) T {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		return res
	}
}

func asyncAddOne(num int) (int, error) {
	time.Sleep(time.Millisecond)
	if num == -2000 {
		panic("set to panic")
	}
	if num < 0 {
		return 0, fmt.Errorf("no negatives numbers allowed for this function")
	}
	return num + 1, nil
}

func asyncCancelable(ctx context.Context, id string) (int, error) {
	if id == "quick" {
		time.Sleep(5 * time.Millisecond)
		return 2000, nil
	}
	for i := range 1000 {
		select {
		case <-ctx.Done():
			return i, fmt.Errorf("cancel " + id)
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
	return 1000, nil
}
