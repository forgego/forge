package main

import (
	"io"
	"net"
	"net/http"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/forgego/forge/config"
)

// The example must shut down gracefully: on SIGINT the in-flight request
// completes and StartWithGracefulShutdown returns nil instead of the process
// dying in log.Fatal(ListenAndServe()) (#294).
func TestEcommerceServerDrainsInFlightRequestOnSignal(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	_ = listener.Close()

	cfg := config.NewConfig()
	cfg.Set("server.host", "127.0.0.1")
	cfg.Set("server.port", strconv.Itoa(port))
	cfg.Set("server.idle_timeout", 7)

	started := make(chan struct{})
	release := make(chan struct{})
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/slow" {
			close(started)
			<-release
		}
		_, _ = io.WriteString(w, "ok")
	})
	srv, err := newEcommerceServer(cfg, handler)
	if err != nil {
		t.Fatal(err)
	}
	if srv.Addr != "127.0.0.1:"+strconv.Itoa(port) || srv.IdleTimeout != 7*time.Second {
		t.Fatalf("server not configured from cfg: addr %q idle %v", srv.Addr, srv.IdleTimeout)
	}

	done := make(chan error, 1)
	go func() { done <- srv.StartWithGracefulShutdown() }()

	base := "http://" + srv.Addr
	deadline := time.Now().Add(5 * time.Second)
	for {
		resp, err := http.Get(base + "/")
		if err == nil {
			_ = resp.Body.Close()
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("server did not start: %v", err)
		}
		time.Sleep(20 * time.Millisecond)
	}

	type result struct {
		body string
		err  error
	}
	slow := make(chan result, 1)
	go func() {
		resp, err := http.Get(base + "/slow")
		if err != nil {
			slow <- result{err: err}
			return
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		slow <- result{body: string(body), err: err}
	}()
	<-started

	if err := syscall.Kill(syscall.Getpid(), syscall.SIGINT); err != nil {
		t.Fatal(err)
	}
	time.Sleep(100 * time.Millisecond)
	close(release)

	got := <-slow
	if got.err != nil || got.body != "ok" {
		t.Fatalf("in-flight request was not drained: body %q err %v", got.body, got.err)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("StartWithGracefulShutdown returned %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("server did not stop after SIGINT")
	}
}
