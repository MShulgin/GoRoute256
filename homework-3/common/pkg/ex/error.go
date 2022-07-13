package ex

import "net/http"

type ErrType int

const (
	BadRequestError     ErrType = 400
	NotFoundError       ErrType = 404
	ConflictError       ErrType = 409
	InternalServerError ErrType = 500
)

type AppError struct {
	Code    ErrType `json:"status"`
	Message string  `json:"message"`
}

func (err *AppError) IsType(t ErrType) bool {
	return err.Code == t
}

func (err *AppError) Error() string {
	return err.Message
}

func (err *AppError) HttpCode() int {
	switch err.Code {
	case BadRequestError:
		return http.StatusBadRequest
	case NotFoundError:
		return http.StatusNotFound
	case ConflictError:
		return http.StatusConflict
	case InternalServerError:
		return http.StatusInternalServerError
	default:
		return 0
	}
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
