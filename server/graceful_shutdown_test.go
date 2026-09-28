package server

import (
	"context"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/forgego/forge/config"
	"github.com/stretchr/testify/require"
)

func freeTCPPort(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	_, port, err := net.SplitHostPort(l.Addr().String())
	require.NoError(t, err)
	require.NoError(t, l.Close())
	return port
}

// TestServeUntilDrainsInFlightRequests checks the shutdown path used by
// StartWithGracefulShutdown: a request that is running when shutdown starts
// completes, the server stops accepting connections, and the call returns nil.
func TestServeUntilDrainsInFlightRequests(t *testing.T) {
	port := freeTCPPort(t)
	settings := &config.Settings{
		App:    config.AppSettings{Env: "test"},
		Server: config.ServerSettings{Host: "127.0.0.1", Port: port, ReadTimeout: 5, WriteTimeout: 5, GracefulTimeout: 5},
	}
	srv, err := NewServer(config.NewConfig(), settings, nil)
	require.NoError(t, err)

	started := make(chan struct{})
	release := make(chan struct{})
	srv.RegisterRoutes(func(r *Router) {
		r.Get("/slow", func(w http.ResponseWriter, _ *http.Request) {
			close(started)
			<-release
			_, _ = io.WriteString(w, "done")
		})
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	serveErr := make(chan error, 1)
	go func() { serveErr <- srv.serveUntil(ctx) }()

	base := "http://127.0.0.1:" + port
	require.Eventually(t, func() bool {
		conn, err := net.Dial("tcp", "127.0.0.1:"+port)
		if err != nil {
			return false
		}
		_ = conn.Close()
		return true
	}, 5*time.Second, 10*time.Millisecond)

	type result struct {
		status int
		body   string
		err    error
	}
	inFlight := make(chan result, 1)
	go func() {
		resp, err := http.Get(base + "/slow")
		if err != nil {
			inFlight <- result{err: err}
			return
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		inFlight <- result{status: resp.StatusCode, body: string(body), err: err}
	}()
	<-started

	cancel()
	// Shutdown closes the listener first; wait until new connections fail.
	require.Eventually(t, func() bool {
		conn, err := net.Dial("tcp", "127.0.0.1:"+port)
		if err != nil {
			return true
		}
		_ = conn.Close()
		return false
	}, 5*time.Second, 10*time.Millisecond)

	select {
	case err := <-serveErr:
		t.Fatalf("serveUntil returned before the in-flight request finished: %v", err)
	default:
	}

	close(release)
	res := <-inFlight
	require.NoError(t, res.err)
	require.Equal(t, http.StatusOK, res.status)
	require.Equal(t, "done", res.body)

	select {
	case err := <-serveErr:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("serveUntil did not return after the in-flight request finished")
	}
}
