package robin

import "log/slog"

type (
	ErrorHandler func(error) ([]byte, int)

	Error struct {
		Message       string
		Code          int
		OriginalError error
	}

	InternalError struct {
		Reason        string
		OriginalError error
	}
)

func (e Error) Error() string {
	return e.Message
}

func (ie InternalError) Error() string {
	return ie.Reason
}

func DefaultErrorHandler(err error) ([]byte, int) {
	var (
		code    = 500
		message = err.Error()
	)

	if e, ok := err.(Error); ok {
		message = e.Message

		if e.Code >= 400 && e.Code < 600 {
			code = e.Code
		}
	} else if e, ok := err.(InternalError); ok {
		message = e.Reason

		slog.Error("An internal error occurred", slog.String("reason", e.Reason), slog.Any("originalError", e.OriginalError.Error()))
	}

	return []byte(message), code
}
