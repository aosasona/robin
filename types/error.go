package types

type (
	CastError struct {
		Expected string
		Actual   string
	}

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

func (ce CastError) Error() string {
	return "failed to cast value, expected `" + ce.Expected + "`, got `" + ce.Actual + "`"
}

func (e Error) Error() string {
	return e.Message
}

func (ie RobinError) Error() string {
	return ie.Reason
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

var (
	_ error = (*CastError)(nil)
	_ error = (*Error)(nil)
	_ error = (*RobinError)(nil)
)
