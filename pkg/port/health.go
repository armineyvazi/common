package port

import (
	"context"
)

type Health interface {
	ServiceName() string
	IsHealthy(ctx context.Context) bool
}
