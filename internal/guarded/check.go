package guarded

import (
	"reflect"

	"go.trulyao.dev/robin/types"
)

// ExpectsPayload returns whether the given parameter expects a payload or not by checking if it is a void type
func ExpectsPayload(param any) types.ExpectedPayloadType {
	if param == nil {
		return types.ExpectedPayloadNone
	}

	if reflect.TypeOf(param).Name() != "_RobinVoid" {
		return types.ExpectedPayloadDecoded
	}

	return types.ExpectedPayloadNone
}
