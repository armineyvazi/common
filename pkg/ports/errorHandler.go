package ports

import "github.com/gofiber/fiber/v2"

type ErrorHandler interface {
	CaptureError(msg string)
	CaptureFiberException(ctx *fiber.Ctx, exception error, params ...any)
	Sync() error
	Write(p []byte) (n int, err error)
}

type ErrorDetails struct {
	HttpStatus  HttpStatus   `json:"-"`
	Status      bool         `json:"status"` // allways false
	Message     string       `json:"message,omitempty"`
	Code        int          `json:"code,omitempty"`
	TrackId     int          `json:"track_id,omitempty"`
	Errors      []FieldError `json:"errors,omitempty"`
	NestedError error        `json:"error,omitempty"`
}

func (e ErrorDetails) Error() string {
	return e.Message
}
