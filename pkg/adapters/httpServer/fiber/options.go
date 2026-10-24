package fiber

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

// Option customises the underlying fiber.Config. Options are applied after
// the adapter's own settings (such as its ErrorHandler), in the order given.
type Option func(*fiber.Config)

// WithTimeouts sets the server's read, write and idle timeouts. A zero value
// leaves that timeout disabled, which is fasthttp's default.
//
// Production servers should set all three: without a read timeout a client
// can hold a connection open indefinitely by sending a request slowly.
func WithTimeouts(read, write, idle time.Duration) Option {
	return func(c *fiber.Config) {
		c.ReadTimeout = read
		c.WriteTimeout = write
		c.IdleTimeout = idle
	}
}

// WithBodyLimit sets the maximum accepted request body size in bytes.
// Requests above the limit are rejected with 413. Fiber's default is 4 MiB.
func WithBodyLimit(bytes int) Option {
	return func(c *fiber.Config) {
		c.BodyLimit = bytes
	}
}

// WithoutStartupMessage disables Fiber's ASCII banner, which otherwise
// interleaves with structured logs on startup.
func WithoutStartupMessage() Option {
	return func(c *fiber.Config) {
		c.DisableStartupMessage = true
	}
}

// WithFiberConfig applies an arbitrary change to the fiber.Config for
// settings without a dedicated option. Changing ErrorHandler replaces the
// adapter's error mapping.
func WithFiberConfig(fn func(*fiber.Config)) Option {
	return Option(fn)
}

func newApp(cfg fiber.Config, opts []Option) *fiber.App {
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return fiber.New(cfg)
}
