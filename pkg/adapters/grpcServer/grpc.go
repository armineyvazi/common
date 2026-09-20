// Package grpcServer provides a ports.GRPCServer adapter backed by
// google.golang.org/grpc. It adds sensible defaults: a recovery interceptor
// that converts handler panics to gRPC Internal errors, and a logging
// interceptor that records method, duration, and outcome via ports.Logger.
package grpcserver

import (
	"context"
	"fmt"
	"net"
	"runtime/debug"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"

	"github.com/armineyvazi/common.git/pkg/ports"
)

// Config tunes the gRPC server. Zero values fall back to gRPC defaults.
type Config struct {
	// MaxRecvMsgSize is the max size in bytes of a received message (default 4 MiB).
	MaxRecvMsgSize int
	// MaxSendMsgSize is the max size in bytes of a sent message (default unlimited).
	MaxSendMsgSize int
	// KeepaliveTime is how often the server sends keepalive pings to idle clients.
	KeepaliveTime time.Duration
	// KeepaliveTimeout is how long the server waits for a keepalive ack.
	KeepaliveTimeout time.Duration
}

type grpcServer struct {
	srv *grpc.Server
	log ports.Logger
}

// New returns a ports.GRPCServer configured with recovery and logging interceptors.
func New(log ports.Logger, cfg Config) ports.GRPCServer {
	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			unaryRecovery(log),
			unaryLogging(log),
		),
		grpc.ChainStreamInterceptor(
			streamRecovery(log),
		),
	}
	if cfg.MaxRecvMsgSize > 0 {
		opts = append(opts, grpc.MaxRecvMsgSize(cfg.MaxRecvMsgSize))
	}
	if cfg.MaxSendMsgSize > 0 {
		opts = append(opts, grpc.MaxSendMsgSize(cfg.MaxSendMsgSize))
	}
	if cfg.KeepaliveTime > 0 {
		opts = append(opts, grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    cfg.KeepaliveTime,
			Timeout: cfg.KeepaliveTimeout,
		}))
	}
	return &grpcServer{
		srv: grpc.NewServer(opts...),
		log: log,
	}
}

// Register registers a gRPC service. desc must be *grpc.ServiceDesc from
// protobuf-generated code; impl is the service implementation.
// Panics if desc is not *grpc.ServiceDesc.
func (s *grpcServer) Register(desc any, impl any) {
	d, ok := desc.(*grpc.ServiceDesc)
	if !ok {
		panic(fmt.Sprintf("grpcServer.Register: desc must be *grpc.ServiceDesc, got %T", desc))
	}
	s.srv.RegisterService(d, impl)
}

// Listen binds to addr and blocks until the server stops.
func (s *grpcServer) Listen(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("grpc listen %s: %w", addr, err)
	}
	s.log.Info("gRPC server listening", "addr", addr)
	return s.srv.Serve(ln)
}

// Shutdown gracefully drains in-flight RPCs. If ctx expires before draining
// completes, it falls back to an immediate hard stop.
func (s *grpcServer) Shutdown(ctx context.Context) {
	stopped := make(chan struct{})
	go func() {
		s.srv.GracefulStop()
		close(stopped)
	}()
	select {
	case <-ctx.Done():
		s.srv.Stop()
	case <-stopped:
	}
}

// unaryRecovery returns a unary interceptor that recovers from panics and
// converts them to a gRPC Internal error, preserving server availability.
func unaryRecovery(log ports.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("grpc panic recovered", "method", info.FullMethod, "panic", r, "stack", string(debug.Stack()))
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()
		return handler(ctx, req)
	}
}

// unaryLogging returns a unary interceptor that logs each RPC with its
// method name, outcome code, and elapsed duration.
func unaryLogging(log ports.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		code := codes.OK
		if err != nil {
			code = status.Code(err)
		}
		log.Info("grpc unary",
			"method", info.FullMethod,
			"code", code.String(),
			"duration_ms", time.Since(start).Milliseconds(),
		)
		return resp, err
	}
}

// streamRecovery returns a stream interceptor that recovers from panics.
func streamRecovery(log ports.Logger) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("grpc stream panic recovered", "method", info.FullMethod, "panic", r, "stack", string(debug.Stack()))
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()
		return handler(srv, ss)
	}
}
