package appErr

import "github.com/armineyvazi/common.git/pkg/ports"

func HandleError(err error) ports.AppError {
	if err == nil {
		return nil
	}

	if m, ok := err.(ports.AppError); ok {
		return m
	}
	e := NewInvalidArgumentErr(err)
	return e
}

func HandleInternalError(err error) ports.AppError {
	if err == nil {
		return nil
	}
	if m, ok := err.(ports.AppError); ok {
		return m
	}
	e := NewInternalErr(err)
	return e
}
