package ports

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type HttpMethod string
type HttpStatus = int
type HttpContext = fiber.Ctx

const (
	Get    HttpMethod = http.MethodGet
	Post   HttpMethod = http.MethodPost
	Put    HttpMethod = http.MethodPut
	Delete HttpMethod = http.MethodDelete
)

const (
	StatusOk                  HttpStatus = http.StatusOK
	StatusPermissionDenied    HttpStatus = http.StatusForbidden
	StatusBadRequest          HttpStatus = http.StatusBadRequest
	StatusUnprocessableEntity HttpStatus = http.StatusUnprocessableEntity
	StatusNotFoundError       HttpStatus = http.StatusNotFound
	StatusAuthorizationError  HttpStatus = http.StatusUnauthorized
	StatusMethodNotAllowed    HttpStatus = http.StatusMethodNotAllowed
	StatusRequestTimeout      HttpStatus = http.StatusRequestTimeout
	StatusPreConditionFailed  HttpStatus = http.StatusPreconditionFailed
	StatusTooManyRequests     HttpStatus = http.StatusTooManyRequests  //429
	StatusFailedDependency    HttpStatus = http.StatusFailedDependency //424
	StatusInternalServerError HttpStatus = http.StatusInternalServerError
	StatusBadGateway          HttpStatus = http.StatusBadGateway
	StatusServiceUnavailable  HttpStatus = http.StatusServiceUnavailable
)

var AppErrorToHttpStatus = map[ErrorCode]HttpStatus{
	InvalidArgumentErrorCode:    StatusUnprocessableEntity,
	PermissionDeniedErrorCode:   StatusPermissionDenied,
	NotFoundErrorCode:           StatusNotFoundError,
	UnauthenticatedErrorCode:    StatusAuthorizationError,
	BadRequestErrorCode:         StatusBadRequest,
	DeadlineExceededErrorCode:   StatusRequestTimeout,
	FailedPreconditionErrorCode: StatusPreConditionFailed,
	TooManyRequestsErrorCode:    StatusTooManyRequests,
	FailedDependencyErrorCode:   StatusFailedDependency,
	InternalErrorCode:           StatusInternalServerError,
	UnknownErrorCode:            StatusInternalServerError,
	MethodNotAllowedErrorCode:   StatusMethodNotAllowed,
}

type Route struct {
	Method  HttpMethod
	Path    string
	Handler func(*HttpContext) error
}

type Request struct {
}

func (r *Request) Normalize() *Request {
	return r
}

type Pagination struct {
	CurrentPage      int   `json:"current_page"`
	TotalCount       int64 `json:"total"`
	TotalPages       int   `json:"total_pages"`
	CurrentPageCount int   `json:"count"`
	PerPage          int   `json:"per_page"`
}

type Meta struct {
	Pagination *Pagination `json:"pagination,omitempty"`
	Include    interface{} `json:"include,omitempty"`
}

type Response struct {
	Status bool        `json:"status"`
	Data   interface{} `json:"data"`
	Meta   *Meta       `json:"meta,omitempty"`
}

type HttpServer interface {
	Listen() error
	SetRouteGroups(groupName string, middlewares []func(ctx *HttpContext) error, routes []Route)
	ActiveApm()
	ActiveLogger()
	ActiveCustomLogger()
	ActiveSwagger(prefix string)
	ActiveSentry()
	ActiveProfiler()
	Test(req *http.Request, msTimeout ...int) (*http.Response, error)
	ActiveRecover()
	ActiveTokenProcessing()
	ActiveTokenValidation(publicKey string)
	ActiveTraceID()
	ActiveSourceNameLog(serviceName string)
	GET(path string, handlers ...fiber.Handler)
	POST(path string, handlers ...fiber.Handler)
	PUT(path string, handlers ...fiber.Handler)
	DELETE(path string, handlers ...fiber.Handler)
	OPTIONS(path string, handlers ...fiber.Handler)
	PATCH(path string, handlers ...fiber.Handler)
	ShutDown() error
	ShutDownWithContext(ctx context.Context) error
	Use(args ...interface{})
	ActiveHealthCheck()
}

type JwtMiddleware interface {
	TokenParser(c *HttpContext) error
	TokenValidation(c *HttpContext) error
}
