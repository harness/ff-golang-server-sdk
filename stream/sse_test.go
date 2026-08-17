package stream

import (
	"context"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"

	"github.com/harness-community/sse/v3"
	"github.com/harness/ff-golang-server-sdk/logger"
)

// TestSubscribeDoesNotLeakGoroutinesOnReconnect reconnects repeatedly against a server that closes
// the stream immediately, and asserts that each subscription cleans up after itself.
func TestSubscribeDoesNotLeakGoroutinesOnReconnect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		w.(http.Flusher).Flush()
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	connected := make(chan struct{})
	disconnected := make(chan error)

	client := &SSEClient{
		client:             sse.NewClient(server.URL),
		logger:             logger.NewNoOpLogger(),
		streamConnected:    connected,
		streamDisconnected: disconnected,
	}

	go func() {
		for {
			select {
			case <-connected:
			case <-ctx.Done():
				return
			}
		}
	}()

	const reconnects = 20

	subscribeOnce := func() {
		out := client.subscribe(ctx, "env", "key")

		select {
		case <-disconnected:
		case <-time.After(10 * time.Second):
			t.Error("timed out waiting for stream to disconnect")
			return
		}

		for range out {
		}
	}

	// Warm up so that goroutines owned by the http client and the sse library are already running
	// when we take the baseline.
	subscribeOnce()
	baseline := settledGoroutines()

	for i := 0; i < reconnects; i++ {
		subscribeOnce()
	}

	// A leaked waiter per reconnect would show up as `reconnects` extra goroutines. Allow a small
	// margin for idle connection handling in the http client.
	const margin = 5
	if got := settledGoroutines(); got > baseline+margin {
		t.Errorf("goroutines grew from %d to %d over %d reconnects, expected no more than %d", baseline, got, reconnects, baseline+margin)
	}
}

// settledGoroutines waits for the goroutine count to stop changing, so that goroutines which are
// still winding down aren't counted as leaks.
func settledGoroutines() int {
	previous := runtime.NumGoroutine()

	for i := 0; i < 50; i++ {
		time.Sleep(20 * time.Millisecond)

		current := runtime.NumGoroutine()
		if current == previous {
			return current
		}
		previous = current
	}

	return previous
}
