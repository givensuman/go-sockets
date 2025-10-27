package emitter

import (
	"sync"
	"testing"
)

func TestOnEmit(t *testing.T) {
	em := &EventEmitter{}
	called := false
	em.On("test", func() {
		called = true
	})
	em.Emit("test")
	if !called {
		t.Error("callback not called")
	}
}

func TestOnEmitWithArgs(t *testing.T) {
	em := &EventEmitter{}
	var received string
	em.On("test", func(s string) {
		received = s
	})
	em.Emit("test", "hello")
	if received != "hello" {
		t.Error("arg not received")
	}
}

func TestMultipleArgs(t *testing.T) {
	em := &EventEmitter{}
	var a, b int
	em.On("test", func(x, y int) {
		a, b = x, y
	})
	em.Emit("test", 1, 2)
	if a != 1 || b != 2 {
		t.Error("args not received")
	}
}

func TestOnce(t *testing.T) {
	em := &EventEmitter{}
	count := 0
	em.Once("test", func() {
		count++
	})
	em.Emit("test")
	em.Emit("test")
	if count != 1 {
		t.Error("once called more than once")
	}
}

func TestOff(t *testing.T) {
	em := &EventEmitter{}
	called := false
	fn := func() {
		called = true
	}
	em.On("test", fn)
	em.Off("test", fn)
	em.Emit("test")
	if called {
		t.Error("callback called after off")
	}
}

func TestConcurrent(t *testing.T) {
	em := &EventEmitter{}
	results := make(chan bool, 100)
	for i := 0; i < 10; i++ {
		em.On("test", func() { results <- true })
	}
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			em.Emit("test")
		}()
	}
	wg.Wait()
	close(results)
	count := 0
	for range results {
		count++
	}
	if count != 100 {
		t.Errorf("expected 100 calls, got %d", count)
	}
}

func TestPanicRecovery(t *testing.T) {
	em := &EventEmitter{}
	em.On("test", func() {
		panic("test panic")
	})
	em.On("test", func() {
		t.Log("second callback called")
	})
	// Should not panic
	em.Emit("test")
}
