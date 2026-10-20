package net_test

import (
	"context"
	"testing"
	"time"

	"AxiomLang/stdlib/net"
)

func TestNet_ListenConnectEcho(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	addr := "127.0.0.1:29876"
	srv, err := net.RunEchoService(ctx, addr)
	if err != nil {
		t.Fatalf("failed to start echo server: %v", err)
	}
	defer srv.Close()

	// Wait briefly for socket binding
	time.Sleep(50 * time.Millisecond)

	client, err := net.Connect(addr, 2*time.Second)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer client.Close()

	msg := []byte("Hello, Axiom Zero-Ambient Network!")
	if _, err := client.Write(msg); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	reply := make([]byte, len(msg))
	n, err := client.Read(reply)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}

	if string(reply[:n]) != string(msg) {
		t.Fatalf("echo mismatch: got %q, expected %q", string(reply[:n]), string(msg))
	}
}
