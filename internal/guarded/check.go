package guarded

import (
	"reflect"

	"go.trulyao.dev/robin/types"
)

// ExpectsPayload returns whether the given parameter expects a payload or not by checking if it is a void type
func ExpectsPayload(param any) bool {
	void := types.Void{}
	return !(reflect.TypeOf(param).Kind() == reflect.TypeOf(void).Kind())
}
