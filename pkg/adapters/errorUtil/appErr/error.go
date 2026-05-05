package appErr

import (
	"context"
	"strconv"
	"time"

	"github.com/armineyvazi/common.git/pkg/adapters/encoder/optimus"
	"github.com/armineyvazi/common.git/pkg/ports"
)

type appErr struct {
	Message          string          `json:"message"`
	Code             ports.ErrorCode `json:"code"`
	TrackId          string          `json:"track_id"`
	Status           bool            `json:"status"`
	Errors           map[ports.ErrorFieldName]ports.ErrorList
	DisableDebug     bool                   `json:"-"`
	ErrType          ports.ErrorType        `json:"-"`
	NestedError      error                  `json:"-"`
	ExtraInformation map[string]interface{} `json:"-"`
	MetaData         interface{}            `json:"-"`
	Detail           string                 `json:"-"`
}

func NewAppError(
	code ports.ErrorCode,
	msg ports.ErrorMessage,
	err error,
) ports.AppError {
	return &appErr{
		NestedError: err,
		Code:        code,
		Message:     msg,
	}
}

func New(e error) (r ports.AppError) {
	r = new(appErr)
	return r.WithError(e)
}

func (e *appErr) WithError(er error) ports.AppError {
	if er == nil {
		return e
	}
	e.NestedError = er
	// Default detail is source error
	return e.WithDetail(e.Error())
}

func (e *appErr) WithCode(code int) ports.AppError {
	e.Code = code
	return e
}

func (e *appErr) WithTrackId(ctx context.Context) ports.AppError {
	// Set trackID if not already set to prevent confusion in logs
	if e.TrackId != "" {
		return e
	}

	trackID, ok := ctx.Value(ports.TraceID{}).(string)
	if ok {
		e.TrackId = trackID
		return e
	}

	e.TrackId = strconv.Itoa(int(optimus.NewOptimusEncoder(optimus.Config{}).
		Encode(uint64(time.Now().Unix()))))
	return e
}

// WithDisableDebug note: with active disable debug error, there is no log for errors
func (e *appErr) WithDisableDebug() ports.AppError {
	e.DisableDebug = true
	return e
}

func (e *appErr) WithMessage(message ports.ErrorMessage) ports.AppError {
	e.Message = message
	return e
}

func (e *appErr) WithDetail(d string) ports.AppError {
	e.Detail = d
	return e
}

func (e *appErr) WithType(t ports.ErrorType) ports.AppError {
	e.ErrType = t
	e.setDefaults()
	return e
}

// WithExtraInfo adds extra information for client in error
// multiple calls of this function adds to the map
// note: it replaces the keys with same name
func (e *appErr) WithExtraInfo(extraInfo map[string]interface{}) ports.AppError {
	if e.ExtraInformation == nil {
		e.ExtraInformation = make(map[string]interface{})
	}

	for k, v := range extraInfo {
		e.ExtraInformation[k] = v
	}

	return e
}

func (e *appErr) WithNestedError(err error) ports.AppError {
	e.NestedError = err
	return e
}

func (e *appErr) WithErrorList(fieldName, errorMessage string) ports.AppError {
	if e.Errors == nil {
		e.Errors = make(map[ports.ErrorFieldName]ports.ErrorList)
	}
	e.Errors[fieldName] = append(e.Errors[fieldName], errorMessage)
	return e
}

// setDefaults set default IDs in case of setting Type
func (e *appErr) setDefaults() {
	if e.Message == "" {
		if defaultMessage, ok := ports.ErrorMessagesMap[e.ErrType]; ok {
			e.Message = defaultMessage
		}
	}
}

func (e *appErr) GetCode() int {
	return e.Code
}

func (e *appErr) GetError() error {
	return e.NestedError
}

func (e *appErr) GetMessage() ports.ErrorMessage {
	return e.Message
}

func (e *appErr) GetTrackId() string {
	return e.TrackId
}

func (e *appErr) GetType() ports.ErrorType {
	return e.ErrType
}

func (e *appErr) GetDetail() string {
	return e.Detail
}

func (e *appErr) GetExtraInfo() map[string]interface{} {
	return e.ExtraInformation
}

func (e *appErr) GetNestedError() error {
	return e.NestedError
}

func (e *appErr) GetErrorLists() map[ports.ErrorFieldName]ports.ErrorList {
	return e.Errors
}

func (e *appErr) IsDebugDisabled() bool {
	return e.DisableDebug
}

func (e *appErr) Error() string {
	var r string
	if e.TrackId != "" {
		r += "trackId: " + e.TrackId + " "
	}
	if e.Message != "" {
		r += "message: " + e.Message + ", "
	}

	if e.Detail != "" {
		r += "detail: " + e.Detail + ", "
	}

	if e.NestedError != nil {
		r += "err: " + e.NestedError.Error()
	}

	return r
}
