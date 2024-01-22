package http

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	"go.elastic.co/apm/module/apmfiber"

	"github.com/armineyvazi/common/pkg/adapter/http/middleware/jwt"
	"github.com/armineyvazi/common/pkg/port"
)

type Fiber struct {
	app     *fiber.App
	address string
}

func New(debug bool, address string, sentry port.ErrorHandler) port.HttpServer {
	if address == "" {
		address = "0.0.0.0:3000"
	}

	return &Fiber{
		app: fiber.New(fiber.Config{
			ErrorHandler: func(ctx *fiber.Ctx, err error) error {
				// save error in sentry
				if sentry != nil {
					sentry.CaptureException(err)
				}

				// if error is a formal known error send it directly
				if e, ok := err.(port.ErrorDetails); ok {
					if !debug {
						e.NestedError = nil
					}
					return ctx.Status(e.HttpStatus).JSON(e)
				}

				var e *fiber.Error
				if errors.As(err, &e) {
					return ctx.Status(e.Code).JSON(port.ErrorDetails{
						Status:  false,
						Message: e.Message,
						Code:    http.StatusInternalServerError,
					})
				}

				if !debug {
					err = nil
				}

				// create a formal error if error is not formal
				return ctx.Status(http.StatusInternalServerError).JSON(port.ErrorDetails{
					Status:      false,
					Message:     strconv.Itoa(http.StatusInternalServerError),
					Code:        http.StatusInternalServerError,
					NestedError: err,
				})
			},
		}),
		// sentry:  sentry,
		address: address,
	}
}

func (f *Fiber) ActiveRecover() {
	f.app.Use(recover.New())
}

func (f *Fiber) ActiveApm() {
	f.app.Use(apmfiber.Middleware())
}

func (f *Fiber) ActiveCustomLogger() {
	f.app.Use(logger.New(logger.Config{
		Format: "${time} | ${status} | ${latency} | ${ip} | ${method} | ${path} | uuid:${header:uuid} | origin:${header:origin} | device:${header:s-device}\n",
	}))
}

func (f *Fiber) ActiveLogger() {
	f.app.Use(logger.New())
}

func (f *Fiber) Listen() error {
	if err := f.app.Listen(f.address); err != nil {
		return err
	}
	return nil
}

func (f *Fiber) ActiveTokenProcessing() {
	f.app.Use(jwt.New())
}

func (f *Fiber) ActiveSwagger(prefix string) {
	f.app.Get(prefix+"/swagger/*", swagger.HandlerDefault)
}

// ShutDown gracefully shuts down the http server
func (f *Fiber) ShutDown() error {
	return f.app.Shutdown()
}

// ShutDownWithContext gracefully shuts down the http server
func (f *Fiber) ShutDownWithContext(ctx context.Context) error {
	return f.app.ShutdownWithContext(ctx)
}

func (f *Fiber) SetRouteGroups(groupName string, middlewares []func(ctx *port.HttpContext) error, routes []port.Route) {
	g := f.app.Group("/" + groupName)

	for _, middleware := range middlewares {
		g.Use(middleware)
	}

	for _, route := range routes {
		g.Add(string(route.Method), route.Path, route.Handler)
	}
}
