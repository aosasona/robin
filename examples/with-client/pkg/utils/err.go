package utils

import (
	"log/slog"

	apperrors "todo/pkg/errors"

	"go.trulyao.dev/robin"
)

func ErrorHandler(err error) (robin.Serializable, int) {
	message := err.Error()
	code := 500

	if e, ok := err.(apperrors.Error); ok {
		code = e.Code
		message = e.Message
	} else if e, ok := err.(robin.Error); ok {
		code = e.Code
		message = "An error occurred"
		slog.Error("An internal error occured", slog.String("message", message))
	}

	return robin.ErrorString(message), code
}
