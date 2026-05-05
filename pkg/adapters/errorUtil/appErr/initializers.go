package appErr

import (
	"github.com/gofiber/fiber/v2"
	"github.com/armineyvazi/common.git/pkg/ports"
)

func NewInternalErr(err error) ports.AppError {
	return New(err).
		WithType(ports.TypeInternal).
		WithCode(ports.InternalErrorCode)
}

func NewInvalidArgumentErr(err error) ports.AppError {
	return New(err).
		WithType(ports.TypeValidation).
		WithCode(ports.InvalidArgumentErrorCode)
}

func NewBadRequestErr(err error) ports.AppError {
	return New(err).
		WithType(ports.TypeValidation).
		WithCode(ports.BadRequestErrorCode)
}

func NewOutOfRangeErr(err error) ports.AppError {
	return New(err).
		WithType(ports.TypeValidation).
		WithCode(ports.OutOfRangeErrorCode)
}

func NewUnimplementedErr(err error) ports.AppError {
	return New(err).
		WithType(ports.TypeUnprocessable).
		WithCode(ports.UnimplementedErrorCode)
}

func NewUnavailableErr(msg string, err error) ports.AppError {
	return New(err).
		WithType(ports.TypeUnprocessable).
		WithCode(ports.UnavailableErrorCode)
}

func NewDataLossErr(err error) ports.AppError {
	return New(err).
		WithType(ports.TypeUnprocessable).
		WithCode(ports.DataLossErrorCode)
}

func NewPermissionDeniedErr(err error) ports.AppError {
	return New(err).
		WithType(ports.TypeForbidden).
		WithCode(ports.PermissionDeniedErrorCode)
}

func NewNotFoundErr(err error) ports.AppError {
	return New(err).
		WithType(ports.TypeNotFound).
		WithCode(ports.NotFoundErrorCode)
}

func NewUnauthenticatedErr(err error) ports.AppError {
	return New(err).
		WithType(ports.TypeUnAuthorized).
		WithCode(ports.UnauthenticatedErrorCode)
}

func NewUnauthorizedErr(err error) ports.AppError {
	return New(err).
		WithType(ports.TypeUnAuthorized).
		WithCode(ports.UnauthorizedErrorCode)
}

func NewFailedPreconditionErr(err error) ports.AppError {
	return New(err).
		WithType(ports.TypeInternal).
		WithCode(ports.FailedPreconditionErrorCode)
}

func NewAbortedErr(err error) ports.AppError {
	return New(err).
		WithType(ports.TypeInternal).
		WithCode(ports.AbortedErrorCode)
}

func NewAlreadyExistsErr(err error) ports.AppError {
	return New(err).
		WithType(ports.TypeDuplicate).
		WithCode(ports.AlreadyExistsErrorCode)
}

func NewResourceExhaustedErr(err error) ports.AppError {
	return New(err).
		WithType(ports.TypeTooManyRequests).
		WithCode(ports.ResourceExhaustedErrorCode)
}

func NewCancelledErrorErr(msg string, err error) ports.AppError {
	return New(err).
		WithType(ports.TypeUnprocessable).
		WithCode(ports.CancelledErrorCode)
}

func NewDeadlineExceededErr(err error) ports.AppError {
	return New(err).
		WithType(ports.TypeInternal).
		WithCode(ports.DeadlineExceededErrorCode)
}

func NewMethodNotAllowedErr(err error) ports.AppError {
	return New(err).
		WithType(ports.TypeMethodNotAllowed).
		WithCode(ports.MethodNotAllowedErrorCode)
}

func NewFiberErr(e *fiber.Error) ports.AppError {
	return NewAppError(e.Code, e.Message, nil)
}
