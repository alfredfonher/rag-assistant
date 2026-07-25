package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

func TestGracefulShutdown(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	srv := &http.Server{Handler: mux}
	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	addr := ln.Addr().String()
	go srv.Serve(ln)
	time.Sleep(50 * time.Millisecond)

	resp, err := http.Get("http://" + addr + "/ping")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
}

func TestSignalChannelWiring(t *testing.T) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGUSR1)
	defer signal.Stop(ch)
	syscall.Kill(syscall.Getpid(), syscall.SIGUSR1)
	select {
	case sig := <-ch:
		if sig != syscall.SIGUSR1 {
			t.Fatalf("expected SIGUSR1, got %v", sig)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout")
	}
}
