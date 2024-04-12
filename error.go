package robin

import "log/slog"

type (
	// TODO: change byte to interface{} and switch on the type to make it more flexible
	ErrorHandler func(error) ([]byte, int)

	Error struct {
		Message string
		Code    int
		Cause   error
		Meta    map[string]interface{}
	}

	RobinError struct {
		Reason        string
		OriginalError error
	}
)

func (e Error) Error() string {
	return e.Message
}

func (ie RobinError) Error() string {
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
	} else if e, ok := err.(RobinError); ok {
		message = e.Reason

		slog.Error("An internal error occurred", slog.String("reason", e.Reason), slog.Any("originalError", e.OriginalError.Error()))
	}

	return []byte(message), code
}

func NewError(message string, code ...int) *Error {
	statucCode := 500
	if len(code) > 0 {
		statucCode = code[0]
	}

	return &Error{Message: message, Code: statucCode}
}

func (e *Error) WithCode(code int) *Error {
	e.Code = code
	return e
}

func (e *Error) WithCause(cause error) *Error {
	if cause == nil {
		return e
	}

	e.Cause = cause
	return e
}

func (e *Error) WithMeta(meta map[string]interface{}) *Error {
	if meta == nil {
		return e
	}

	e.Meta = meta
	return e
}

var _ error = (*Error)(nil)
