package httpError

import (
	"encoding/json"

	app_err "github.com/armineyvazi/common.git/pkg/adapters/errorUtil/appErr"
	"github.com/armineyvazi/common.git/pkg/ports"
)

type httpErr struct {
	HttpStatus     ports.HttpStatus `json:"-"`
	ports.AppError `json:"-"`
}

func (e *httpErr) MarshalJSON() ([]byte, error) {
	errorListMap := e.AppError.GetErrorLists()
	errorList := make([]struct {
		FieldName string   `json:"field_name"`
		Errors    []string `json:"errors"`
	}, 0, len(errorListMap))

	for field, errs := range errorListMap {
		errorList = append(errorList, struct {
			FieldName string   `json:"field_name"`
			Errors    []string `json:"errors"`
		}{
			FieldName: field,
			Errors:    errs,
		})
	}

	return json.Marshal(&struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		TrackID string      `json:"track_id"`
		Status  bool        `json:"status"`
		Errors  interface{} `json:"errors"`
	}{
		Code:    e.AppError.GetCode(),
		Message: e.AppError.GetMessage(),
		TrackID: e.AppError.GetTrackId(),
		Status:  false,
		Errors:  errorList,
	})
}

func New(
	appError ports.AppError,
	httpStatus ports.HttpStatus,
) ports.HttpError {
	return &httpErr{
		AppError:   appError,
		HttpStatus: httpStatus,
	}
}

func (e *httpErr) Error() string {
	return e.AppError.Error()
}

func (e *httpErr) GetHttpStatus() int {
	return e.HttpStatus
}

// AddToMap adds new errors to the map
//
// This function is used to add new errors to the map of app errors to http status codes.
// The map is used to map app errors to http status codes.
//
// Note: This function should be called at initiation time of the application.
func AddToMap(newErrs map[int]int) {
	for appErr, httpErr := range newErrs {
		ports.AppErrorToHttpStatus[appErr] = httpErr
	}
}

func MapToHttpErr(err error) ports.HttpError {
	appErr, ok := err.(ports.AppError)
	if !ok {
		appErr = app_err.NewInternalErr(app_err.NewInternalErr(nil))
	}
	code := appErr.GetCode()
	httpStatusCode, ok := ports.AppErrorToHttpStatus[code]
	if !ok {
		httpStatusCode = ports.StatusInternalServerError
	}
	return New(appErr, httpStatusCode)
}

func NewBadRequestErr(err error) ports.HttpError {
	return New(app_err.NewBadRequestErr(err), ports.StatusBadRequest)
}

func NewUnauthorizedErr(err error) ports.HttpError {
	return New(app_err.NewUnauthorizedErr(err), ports.StatusAuthorizationError)
}

func NewForbiddenErr(err error) ports.HttpError {
	return New(app_err.NewPermissionDeniedErr(err), ports.StatusPermissionDenied)
}

func NewNotFoundErr(err error) ports.HttpError {
	return New(app_err.NewNotFoundErr(err), ports.StatusNotFoundError)
}

func NewInternalServerErrorErr(err error) ports.HttpError {
	return New(app_err.NewInternalErr(err), ports.StatusInternalServerError)
}

func NewServiceUnavailableErr(err error) ports.HttpError {
	return New(app_err.NewInternalErr(err), ports.StatusServiceUnavailable)
}

func NewMethodNotAllowedErr(err error) ports.HttpError {
	return New(app_err.NewBadRequestErr(err), ports.StatusMethodNotAllowed)
}

func NewRequestTimeoutErr(err error) ports.HttpError {
	return New(app_err.NewDeadlineExceededErr(err), ports.StatusRequestTimeout)
}

func NewPreConditionFailedErr(err error) ports.HttpError {
	return New(app_err.NewFailedPreconditionErr(err), ports.StatusPreConditionFailed)
}

func NewBadGatewayErr(err error) ports.HttpError {
	return New(app_err.NewInternalErr(err), ports.StatusBadGateway)
}

func NewUnknownErr(err error) ports.HttpError {
	return New(app_err.NewInternalErr(err), ports.StatusInternalServerError)
}
