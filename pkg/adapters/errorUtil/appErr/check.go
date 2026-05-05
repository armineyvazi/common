package appErr

import "github.com/armineyvazi/common.git/pkg/ports"

func (e *appErr) Is(errType ports.ErrorType) bool {
	return e.GetType() == errType
}

func IsErrorType(err error, errType ports.ErrorType) bool {
	if e, ok := err.(ports.AppError); ok {
		return e.Is(errType)
	}
	return false
}

func IsNotFoundError(err error) bool {
	e, ok := err.(ports.AppError)
	if !ok {
		return false
	}
	return e.GetType() == ports.TypeNotFound
}

func IsDuplicate(err error) bool {
	e, ok := err.(ports.AppError)
	if !ok {
		return false
	}
	return e.GetType() == ports.TypeDuplicate
}

func IsInternalError(err error) bool {
	e, ok := err.(ports.AppError)
	if !ok {
		return false
	}
	return e.GetType() == ports.TypeInternal
}

func IsUnauthorizedError(err error) bool {
	e, ok := err.(ports.AppError)
	if !ok {
		return false
	}
	return e.GetType() == ports.TypeUnAuthorized
}

func IsForbiddenError(err error) bool {
	e, ok := err.(ports.AppError)
	if !ok {
		return false
	}
	return e.GetType() == ports.TypeForbidden
}

func IsValidationError(err error) bool {
	e, ok := err.(ports.AppError)
	if !ok {
		return false
	}
	return e.GetType() == ports.TypeValidation
}
