package port

import (
	"context"

	"github.com/gofiber/fiber/v2"
)

type (
	HttpMethod   string
	HttpStatus   = int
	HttpContext  = fiber.Ctx
	ErrorCode    = int
	ErrorMessage = string
)

type Route struct {
	Method  HttpMethod
	Path    string
	Handler func(*HttpContext) error
}

type HttpServer interface {
	Listen() error
	SetRouteGroups(groupName string, middlewares []func(ctx *HttpContext) error, routes []Route)
	ActiveApm()
	ActiveLogger()
	ActiveCustomLogger()
	ActiveSwagger(prefix string)
	// ActiveSentry()
	ActiveRecover()
	ActiveTokenProcessing()
	ShutDown() error
	ShutDownWithContext(ctx context.Context) error
}
