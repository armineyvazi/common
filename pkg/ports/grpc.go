package ports

import "context"

// GRPCServer manages the lifecycle of a gRPC server.
// Register must be called before Listen.
type GRPCServer interface {
	// Register registers a gRPC service. desc must be *grpc.ServiceDesc
	// from the generated protobuf code; impl is the service implementation.
	Register(desc any, impl any)
	// Listen binds to addr and blocks until the server stops or returns an error.
	Listen(addr string) error
	// Shutdown gracefully drains in-flight RPCs then stops the server.
	// It falls back to a hard stop if ctx is cancelled before draining completes.
	Shutdown(ctx context.Context)
}
