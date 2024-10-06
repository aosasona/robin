package apperrors

import (
	"log/slog"

	"go.trulyao.dev/robin"
)

func ErrorHandler(err error) (robin.Serializable, int) {
	message := err.Error()
	code := 500

	switch err := err.(type) {
	case Error:
		code = err.Code
		message = err.Message

	case robin.Error:
		code = err.Code
		message = "An error occurred"
		slog.Error("A robin error occured, this might be a bug", slog.String("message", message))

	default:
		slog.Error("An internal error occured", slog.String("message", message))
	}

	if e, ok := err.(Error); ok {
		code = e.Code
		message = e.Message
	} else if e, ok := err.(robin.Error); ok {
		code = e.Code
		message = "An error occurred"
		slog.Error("An internal error occured", slog.String("message", message))
	} else {
		slog.Error("An internal error occured", slog.String("message", message))
	}

	return robin.ErrorString(message), code
}

type Error struct {
	Message string
	Code    int
}

func New(code int, message string) Error {
	return Error{
		Message: message,
		Code:    code,
	}
}

func (e Error) Error() string {
	return e.Message
}
