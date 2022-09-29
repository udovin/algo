package futures

import (
	"context"
	"testing"
	"time"
)

func TestNewFuture(t *testing.T) {
	{
		future, setResult := New[string]()
		setResult("success", nil)
		result, err := future.Get(context.Background())
		if err != nil {
			t.Fatal("Error:", err)
		}
		if result != "success" {
			t.Fatalf("result has value %q", result)
		}
	}
	{
		future, setResult := New[string]()
		setResult("success", nil)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if result, err := future.Get(ctx); err != nil {
			t.Fatal("Error:", err)
		} else if result != "success" {
			t.Fatalf("result has value %q", result)
		}
	}
	{
		future, _ := New[string]()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if _, err := future.Get(ctx); err != context.Canceled {
			t.Fatal("Error:", err)
		}
	}
	{
		future, setResult := New[string]()
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		go func() {
			setResult("success", nil)
		}()
		if result, err := future.Get(ctx); err != nil {
			t.Fatal("Error:", err)
		} else if result != "success" {
			t.Fatalf("result has value %q", result)
		}
	}
	{
		future, setResult := New[string]()
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		go func() {
			time.Sleep(2 * time.Millisecond)
			setResult("success", nil)
		}()
		if _, err := future.Get(ctx); err != context.DeadlineExceeded {
			t.Fatal("Error:", err)
		}
	}
}
