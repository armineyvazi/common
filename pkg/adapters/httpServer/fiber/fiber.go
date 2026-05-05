package fiber

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/template/html/v2"
	app_err "github.com/armineyvazi/common.git/pkg/adapters/errorUtil/appErr"
	"github.com/armineyvazi/common.git/pkg/adapters/errorUtil/httpError"

	sentryfiber "github.com/getsentry/sentry-go/fiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	"github.com/armineyvazi/common.git/pkg/adapters/httpServer/fiber/middleware/jwt"
	"github.com/armineyvazi/common.git/pkg/ports"
	"go.elastic.co/apm/module/apmfiber/v2"
)

const defaultAddress = "0.0.0.0:3000"

type View struct {
	Dir       string
	Extension string
}

type FiberHttpServer struct {
	app     *fiber.App
	debug   bool
	sentry  ports.ErrorHandler
	address string
	uuidGen ports.UUID
}

func (fh *FiberHttpServer) ActiveRecover() {
	fh.app.Use(recover.New())
}

func NewFiberWithRenderHtml(debug bool, address string, uuidGen ports.UUID, sentry ports.ErrorHandler, view View) ports.HttpServer {
	if address == "" {
		address = defaultAddress
	}

	engine := html.New(view.Dir, view.Extension)

	return &FiberHttpServer{
		app: fiber.New(fiber.Config{
			ErrorHandler: func(ctx *fiber.Ctx, err error) error {
				// save error in sentry
				if sentry != nil {
					sentry.CaptureFiberException(ctx, err)
				}

				// if error is a formal known error send it directly
				if e, ok := err.(ports.AppError); ok {
					httpErr := httpError.MapToHttpErr(e)
					return ctx.Status(httpErr.GetHttpStatus()).JSON(httpErr)
				}

				var e *fiber.Error
				if errors.As(err, &e) {
					return ctx.Status(e.Code).JSON(httpError.NewInternalServerErrorErr(err))
				}

				if !debug {
					err = nil
				}

				// create a formal error if error is not formal
				return ctx.Status(ports.StatusInternalServerError).JSON(httpError.NewInternalServerErrorErr(err))
			},
			Views: engine,
		}),
		sentry:  sentry,
		address: address,
		uuidGen: uuidGen,
	}
}

func New(debug bool, address string, uuidGen ports.UUID, sentry ports.ErrorHandler) ports.HttpServer {
	if address == "" {
		address = defaultAddress
	}

	return &FiberHttpServer{
		app: fiber.New(fiber.Config{
			ErrorHandler: func(ctx *fiber.Ctx, err error) error {
				// save error in sentry
				if sentry != nil {
					sentry.CaptureFiberException(ctx, err)
				}

				// if error is a formal known error send it directly
				if e, ok := err.(ports.ErrorDetails); ok {
					if !debug {
						e.NestedError = nil
					}
					return ctx.Status(e.HttpStatus).JSON(e)
				}

				var e *fiber.Error
				if errors.As(err, &e) {
					return ctx.Status(e.Code).JSON(ports.ErrorDetails{
						Status:  false,
						Message: e.Message,
						Code:    e.Code,
					})
				}

				if !debug {
					err = nil
				}

				// create a formal error if error is not formal
				return ctx.Status(ports.InternalErrorCode).JSON(ports.ErrorDetails{
					Status:      false,
					Message:     ports.InternalErrorMessage,
					Code:        ports.InternalErrorCode,
					NestedError: err,
				})
			},
		}),
		sentry:  sentry,
		address: address,
		uuidGen: uuidGen,
	}
}

func (fh *FiberHttpServer) ActiveApm() {
	fh.app.Use(apmfiber.Middleware())
}

func (fh *FiberHttpServer) ActiveLogger() {
	fh.app.Use(logger.New())
}

func escapeString(s string) string {
	return strings.ReplaceAll(s, `"`, `\"`)
}

func (fh *FiberHttpServer) ActiveCustomLogger() {
	fh.app.Use(logger.New(logger.Config{
		CustomTags: map[string]logger.LogFunc{
			"user-id": func(output logger.Buffer, c *fiber.Ctx, data *logger.Data, extraParam string) (int, error) {
				userIdS, _ := c.Locals("user_id").(string)
				return output.WriteString(userIdS)
			},
			"custom-ip": func(output logger.Buffer, c *fiber.Ctx, data *logger.Data, extraParam string) (int, error) {
				ip := ""
				if len(c.IPs()) > 0 {
					ip = c.IPs()[0]
				}
				return output.WriteString(ip)
			},
			"custom-body": func(output logger.Buffer, c *fiber.Ctx, data *logger.Data, extraParam string) (int, error) {
				body := string(c.Body())
				return output.WriteString(escapeString(body))
			},
		},
		DisableColors: true,
		TimeZone:      "Asia/Tehran",
		TimeFormat:    time.DateTime,
		Format: `{"time":"${time}", "status":${status}, "latency":"${latency}", "path":"${path}", "method":"${method}", ` +
			`"params":"${queryParams}", "body":"${custom-body}", "origin":"${header:origin}", "device":"${header:s-device}", ` +
			`"user_id":"${user-id}", ip":"${custom-ip}", "uuid":"${header:uuid}", "trace_id":"${respHeader:trace_id}"}` + "\n",
	}))
}

func (fh *FiberHttpServer) Use(args ...interface{}) {
	fh.app.Use(args...)
}

func (fh *FiberHttpServer) ActiveProfiler() {
	fh.app.Use(pprof.New())
}

func (fh *FiberHttpServer) ActiveTraceID() {
	fh.app.Use(func(c *fiber.Ctx) error {
		uuid, _ := fh.uuidGen.GenV4() // TODO handle uuid error
		ctxWithUUID := context.WithValue(c.UserContext(), ports.TraceID{}, uuid)
		c.SetUserContext(ctxWithUUID)
		c.Set(ports.TraceIDKey, uuid)
		return c.Next()
	})
}

func (fh *FiberHttpServer) ActiveSourceNameLog(service string) {
	fh.app.Use(func(c *fiber.Ctx) error {
		ctxWithSourceName := context.WithValue(
			c.UserContext(),
			ports.SourceName{},
			service)
		c.SetUserContext(ctxWithSourceName)
		c.Set(ports.SourceNameKey, service)
		return c.Next()
	})
}

func (fh *FiberHttpServer) ActiveSwagger(prefix string) {
	fh.app.Get(prefix+"/swagger/*", swagger.HandlerDefault)
}

func (fh *FiberHttpServer) GET(path string, handlers ...fiber.Handler) {
	fh.app.Get(path, handlers...)
}

func (fh *FiberHttpServer) POST(path string, handlers ...fiber.Handler) {
	fh.app.Post(path, handlers...)
}

func (fh *FiberHttpServer) PUT(path string, handlers ...fiber.Handler) {
	fh.app.Put(path, handlers...)
}

func (fh *FiberHttpServer) DELETE(path string, handlers ...fiber.Handler) {
	fh.app.Delete(path, handlers...)
}

func (fh *FiberHttpServer) OPTIONS(path string, handlers ...fiber.Handler) {
	fh.app.Options(path, handlers...)
}

func (fh *FiberHttpServer) PATCH(path string, handlers ...fiber.Handler) {
	fh.app.Patch(path, handlers...)
}

func (fh *FiberHttpServer) ActiveSentry() {
	fh.app.Use(sentryfiber.New(sentryfiber.Options{}))
}

func (fh *FiberHttpServer) ActiveTokenProcessing() {
	fh.app.Use(jwt.New(""))
}

func (fh *FiberHttpServer) ActiveTokenValidation(publicKey string) {
	if publicKey == "" {
		panic("public key is empty")
	}
	fh.app.Use(jwt.New(publicKey))
}

func (fh *FiberHttpServer) Listen() error {
	if err := fh.app.Listen(fh.address); err != nil {
		return err
	}
	return nil
}

// ShutDown gracefully shuts down the http server
func (fh *FiberHttpServer) ShutDown() error {
	return fh.app.Shutdown()
}

// ShutDownWithContext gracefully shuts down the http server
func (fh *FiberHttpServer) ShutDownWithContext(ctx context.Context) error {
	return fh.app.ShutdownWithContext(ctx)
}

func (fh *FiberHttpServer) Test(req *http.Request, msTimeout ...int) (*http.Response, error) {
	return fh.app.Test(req, msTimeout...)
}

func NewFiberWithErrorModel(
	debug bool,
	logger ports.LoggerWithTraceID,
	address string,
	uuidGen ports.UUID,
	sentry ports.ErrorHandler,
) ports.HttpServer {
	return &FiberHttpServer{
		app: fiber.New(fiber.Config{
			ErrorHandler: func(ctx *fiber.Ctx, err error) (e error) {
				var appErr ports.AppError
				var httpErr ports.HttpError

				switch ProcessableError := err.(type) {
				case ports.AppError:
					appErr = ProcessableError
				case *fiber.Error:

					var e *fiber.Error
					if errors.As(err, &e) {
						appErr = app_err.NewFiberErr(e)
						httpErr = httpError.New(appErr, e.Code)

					} else {
						appErr = app_err.NewInternalErr(err).
							WithDetail("*fiber.Error: could not parse fiber.Error")
					}

				default:
					// create a formal error if error is not formal
					appErr = app_err.HandleInternalError(err).
						WithDetail("error is not formal")
				}

				if httpErr == nil {
					httpErr = httpError.MapToHttpErr(
						appErr.WithTrackId(ctx.UserContext()))
				}

				if debug || !appErr.IsDebugDisabled() {
					sentry.CaptureFiberException(ctx, httpErr, httpErr.GetMetaData(ctx), httpErr.GetExtraInfo())
				}
				return ctx.Status(httpErr.GetHttpStatus()).
					JSON(httpErr)
			},
		}),
		sentry:  sentry,
		address: address,
		uuidGen: uuidGen,
	}
}

func (fh *FiberHttpServer) ActiveHealthCheck() {
	fh.SetRouteGroups("v1/health", nil, []ports.Route{
		{
			Method: ports.Get,
			Path:   "/check",
			Handler: func(context *ports.HttpContext) error {
				return context.JSON(ports.Response{
					Status: true,
					Data:   "alive",
				})
			},
		},
		{
			Method: ports.Get,
			Path:   "/alive",
			Handler: func(context *ports.HttpContext) error {
				return context.JSON(ports.Response{
					Status: true,
					Data:   "alive",
				})
			},
		},
	})
}
