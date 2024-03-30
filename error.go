package robin

import (
	"reflect"
	"strings"
)

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

func InvalidTypes(expected, got any) error {
	expectedType := reflect.TypeOf(expected).Name()
	gotType := reflect.TypeOf(got).String()

	expectedType = strings.ReplaceAll(expectedType, "\"", "")
	gotType = strings.ReplaceAll(gotType, "\"", "")

	return InternalError{Reason: "Invalid types, expected " + expectedType + ", got " + gotType}
}
