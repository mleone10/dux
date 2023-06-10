package dux_test

import (
	"context"
	"testing"
	"time"

	"github.com/mleone10/dux"
)

type mockWatcher struct {
	Sig <-chan bool
}

func (m mockWatcher) Watch(ctx context.Context) {
	<-m.Sig
	return
}

type mockCloseable struct{}

func (m mockCloseable) Close() {}

func TestFuncEngine(t *testing.T) {
	actual, expected := 0, 4
	sigChan := make(chan bool)

	fe := dux.FuncEngine{
		Func: func() dux.Closer {
			actual++
			return mockCloseable{}
		},
		Watcher: mockWatcher{
			Sig: sigChan,
		},
	}

	cancelCtx, cancelFn := context.WithCancel(context.Background())
	go fe.Run(cancelCtx)

	for i := 0; i < expected-1; i++ {
		sigChan <- true
	}
	cancelFn()

	if actual != expected {
		t.Errorf("Expected %v invocations, got %v", expected, actual)
	}
}

func TestTimeWatcher(t *testing.T) {
	expected := time.Millisecond * 200
	tw := dux.TimeWatcher{
		Delay: expected,
	}

	start := time.Now()

	tw.Watch(context.Background())

	end := time.Now()

	dur := end.Sub(start)
	if dur < expected {
		t.Errorf("TimeWatcher unblocked after %v, expected at least %v", dur, expected)
	}
}
