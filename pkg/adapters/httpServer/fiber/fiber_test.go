package fiber_test

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	fiberserver "github.com/armineyvazi/common/pkg/adapters/httpServer/fiber"
	zaplogger "github.com/armineyvazi/common/pkg/adapters/logger/zap"
	"github.com/armineyvazi/common/pkg/adapters/uuid"
	"github.com/armineyvazi/common/pkg/ports"
	"github.com/gofiber/fiber/v2"
)

// TestNewFiberWithErrorModel_NilErrorHandler is a regression test: a nil
// error handler used to panic on the first failed request.
func TestNewFiberWithErrorModel_NilErrorHandler(t *testing.T) {
	srv := fiberserver.NewFiberWithErrorModel(false, zaplogger.New(ports.Error, nil), ":0", uuid.New(), nil)

	resp, err := srv.Test(httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/missing", nil))
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

// serve starts srv on a free loopback port and returns its base URL. The
// in-memory Test() helper cannot be used for body limits: fasthttp reports an
// oversized body as a client-side error there instead of answering 413.
func serve(t *testing.T, newServer func(addr string) ports.HttpServer) string {
	t.Helper()
	ln, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve port: %v", err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()

	srv := newServer(addr)
	srv.POST("/echo", func(c *fiber.Ctx) error { return c.SendStatus(http.StatusNoContent) })
	go func() { _ = srv.Listen() }()
	t.Cleanup(func() { _ = srv.ShutDown() })

	deadline := time.Now().Add(2 * time.Second)
	for {
		conn, err := (&net.Dialer{}).DialContext(context.Background(), "tcp", addr)
		if err == nil {
			_ = conn.Close()
			return "http://" + addr
		}
		if time.Now().After(deadline) {
			t.Fatalf("server did not start on %s: %v", addr, err)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func post(t *testing.T, url, body string) int {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url+"/echo", strings.NewReader(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	return resp.StatusCode
}

func TestOptions(t *testing.T) {
	opts := []fiberserver.Option{
		fiberserver.WithBodyLimit(16),
		fiberserver.WithTimeouts(time.Second, time.Second, time.Second),
		fiberserver.WithoutStartupMessage(),
	}
	constructors := map[string]func(addr string) ports.HttpServer{
		"New": func(addr string) ports.HttpServer {
			return fiberserver.New(false, addr, uuid.New(), nil, opts...)
		},
		"NewFiberWithErrorModel": func(addr string) ports.HttpServer {
			return fiberserver.NewFiberWithErrorModel(false, zaplogger.New(ports.Error, nil), addr, uuid.New(), nil, opts...)
		},
		"NewFiberWithRenderHtml": func(addr string) ports.HttpServer {
			return fiberserver.NewFiberWithRenderHtml(false, addr, uuid.New(), nil,
				fiberserver.View{Dir: t.TempDir(), Extension: ".html"}, opts...)
		},
	}
	for name, newServer := range constructors {
		t.Run(name, func(t *testing.T) {
			url := serve(t, newServer)
			if got := post(t, url, "small"); got != http.StatusNoContent {
				t.Errorf("within limit: status = %d, want 204", got)
			}
			if got := post(t, url, strings.Repeat("x", 64)); got != http.StatusRequestEntityTooLarge {
				t.Errorf("over limit: status = %d, want 413", got)
			}
		})
	}
}
