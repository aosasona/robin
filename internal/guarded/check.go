package guarded

import (
	"reflect"
)

// ExpectsPayload returns whether the given parameter expects a payload or not by checking if it is a void type
func ExpectsPayload(param any) bool {
	return reflect.TypeOf(param).Name() != "_RobinVoid"
}
