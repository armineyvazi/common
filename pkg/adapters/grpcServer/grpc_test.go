package grpcserver_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	grpcserver "github.com/armineyvazi/common.git/pkg/adapters/grpcServer"
)

// nopLogger satisfies ports.Logger without importing the full adapter stack.
type nopLogger struct{}

func (n *nopLogger) Info(msg string, params ...any)  {}
func (n *nopLogger) Warn(msg string, params ...any)  {}
func (n *nopLogger) Error(msg string, params ...any) {}
func (n *nopLogger) Panic(msg string, params ...any) {}

func freePort(t *testing.T) string {
	t.Helper()
	lc := net.ListenConfig{}
	ln, err := lc.Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("find free port: %v", err)
	}
	addr := ln.Addr().String()
	if err := ln.Close(); err != nil {
		t.Fatalf("close listener: %v", err)
	}
	return addr
}

func TestNew_ListenAndShutdown(t *testing.T) {
	addr := freePort(t)
	srv := grpcserver.New(&nopLogger{}, grpcserver.Config{})

	errCh := make(chan error, 1)
	go func() { errCh <- srv.Listen(addr) }()

	// Give the server a moment to bind.
	time.Sleep(50 * time.Millisecond)

	// Verify the server accepts connections.
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	_ = conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	srv.Shutdown(ctx)

	select {
	case err := <-errCh:
		// grpc.Server.Serve returns nil on graceful stop.
		if err != nil {
			t.Fatalf("Listen returned unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server did not shut down within timeout")
	}
}

func TestNew_RegisterWrongType_Panics(t *testing.T) {
	srv := grpcserver.New(&nopLogger{}, grpcserver.Config{})
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when registering wrong desc type")
		}
	}()
	srv.Register("not-a-service-desc", nil)
}

func TestNew_InvalidAddr(t *testing.T) {
	srv := grpcserver.New(&nopLogger{}, grpcserver.Config{})
	err := srv.Listen(fmt.Sprintf("invalid-host:%d", 99999))
	if err == nil {
		t.Fatal("expected error for invalid address")
	}
}
