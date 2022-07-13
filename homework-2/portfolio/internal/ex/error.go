package ex

import (
	"google.golang.org/grpc/codes"
)

type ErrType int

const (
	BadRequestError     ErrType = 400
	NotFoundError       ErrType = 404
	ConflictError       ErrType = 409
	InternalServerError ErrType = 500
)

type AppError struct {
	Code    ErrType
	Message string
}

func (err *AppError) IsType(t ErrType) bool {
	return err.Code == t
}

func (err *AppError) Error() string {
	return err.Message
}

func (err *AppError) GrpcCode() codes.Code {
	switch err.Code {
	case NotFoundError:
		return codes.NotFound
	case ConflictError:
		return codes.InvalidArgument
	case InternalServerError:
		return codes.Internal
	case BadRequestError:
		return codes.InvalidArgument
	}
	return codes.Unknown
}

func NewUnexpectedError(msg string) *AppError {
	return &AppError{
		Code:    InternalServerError,
		Message: msg,
	}
}

func NewNotFoundError(msg string) *AppError {
	return &AppError{
		Code:    NotFoundError,
		Message: msg,
	}
}

func NewConflictError(msg string) *AppError {
	return &AppError{
		Code:    ConflictError,
		Message: msg,
	}
}

func NewBadRequestError(msg string) *AppError {
	return &AppError{
		Code:    BadRequestError,
		Message: msg,
	}
}
