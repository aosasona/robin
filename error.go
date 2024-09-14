package robin

import (
	"log/slog"

	"go.trulyao.dev/robin/types"
)

type (
	ErrorHandler func(error) ([]byte, int)
)

func DefaultErrorHandler(err error) ([]byte, int) {
	var (
		code    = 500
		message = err.Error()
	)

	switch e := err.(type) {
	case types.Error:
		message = e.Message
		if e.Code >= 400 && e.Code < 600 {
			code = e.Code
		}

	case types.RobinError:
		message = e.Reason
		slog.Error("An internal error occurred", slog.String("reason", e.Reason), slog.Any("originalError", e.OriginalError.Error()))

	default:
		message = e.Error()
	}

	return []byte(message), code
}
