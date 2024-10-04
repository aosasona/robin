package robin

import (
	"encoding/json"
	"log/slog"

	"go.trulyao.dev/robin/types"
)

type (
	// A type that can be marshalled to JSON or simply a string
	Serializable interface {
		json.Marshaler
	}

	ErrorHandler func(error) (Serializable, int)

	ErrorString string
)

func DefaultErrorHandler(err error) (Serializable, int) {
	var (
		code    = 500
		message string
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

	return ErrorString(message), code
}

func (e ErrorString) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(e))
}
