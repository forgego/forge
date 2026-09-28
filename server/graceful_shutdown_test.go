package server

import (
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"syscall"
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

// TestStartWithGracefulShutdownSecondSignalKills runs a server in a child
// process with a request that never finishes, sends SIGTERM to start the
// drain, then a second SIGTERM. The second signal must kill the child at once
// instead of being swallowed until server.graceful_timeout expires.
func TestStartWithGracefulShutdownSecondSignalKills(t *testing.T) {
	if os.Getenv("FORGE_TEST_SIGNAL_CHILD") == "1" {
		runSignalChild()
		return
	}
	if runtime.GOOS == "windows" {
		t.Skip("POSIX signals only")
	}

	port := freeTCPPort(t)
	cmd := exec.Command(os.Args[0], "-test.run=^TestStartWithGracefulShutdownSecondSignalKills$")
	cmd.Env = append(os.Environ(), "FORGE_TEST_SIGNAL_CHILD=1", "FORGE_TEST_SIGNAL_PORT="+port)
	require.NoError(t, cmd.Start())
	exited := make(chan error, 1)
	go func() { exited <- cmd.Wait() }()
	defer func() { _ = cmd.Process.Kill() }()

	base := "http://127.0.0.1:" + port
	require.Eventually(t, func() bool {
		conn, err := net.Dial("tcp", "127.0.0.1:"+port)
		if err != nil {
			return false
		}
		_ = conn.Close()
		return true
	}, 10*time.Second, 20*time.Millisecond)

	// Hold a request open so the drain cannot finish on its own.
	go func() {
		resp, err := http.Get(base + "/hang")
		if err == nil {
			_ = resp.Body.Close()
		}
	}()
	time.Sleep(200 * time.Millisecond)

	require.NoError(t, cmd.Process.Signal(syscall.SIGTERM))
	// The drain has started once the listener is closed.
	require.Eventually(t, func() bool {
		conn, err := net.Dial("tcp", "127.0.0.1:"+port)
		if err != nil {
			return true
		}
		_ = conn.Close()
		return false
	}, 10*time.Second, 20*time.Millisecond)

	select {
	case err := <-exited:
		t.Fatalf("child exited before the second signal: %v", err)
	default:
	}

	require.NoError(t, cmd.Process.Signal(syscall.SIGTERM))
	select {
	case err := <-exited:
		var exitErr *exec.ExitError
		require.ErrorAs(t, err, &exitErr)
		status, ok := exitErr.Sys().(syscall.WaitStatus)
		require.True(t, ok)
		require.True(t, status.Signaled(), "child should die from the second signal, got %v", err)
		require.Equal(t, syscall.SIGTERM, status.Signal())
	case <-time.After(5 * time.Second):
		t.Fatal("second SIGTERM was swallowed during the graceful drain")
	}
}

func runSignalChild() {
	settings := &config.Settings{
		App:    config.AppSettings{Env: "test"},
		Server: config.ServerSettings{Host: "127.0.0.1", Port: os.Getenv("FORGE_TEST_SIGNAL_PORT"), ReadTimeout: 60, WriteTimeout: 60, GracefulTimeout: 60},
	}
	srv, err := NewServer(config.NewConfig(), settings, nil)
	if err != nil {
		os.Exit(3)
	}
	srv.RegisterRoutes(func(r *Router) {
		r.Get("/hang", func(_ http.ResponseWriter, req *http.Request) {
			<-make(chan struct{})
		})
	})
	_ = srv.StartWithGracefulShutdown()
	os.Exit(0)
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
	signalsReleased := make(chan struct{})
	go func() { serveErr <- srv.serveUntil(ctx, func() { close(signalsReleased) }) }()

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
	// Signal handling must be released while the request is still draining,
	// so a second Ctrl+C is not swallowed.
	select {
	case <-signalsReleased:
	case <-time.After(5 * time.Second):
		t.Fatal("serveUntil did not release signal handling before draining")
	}
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
