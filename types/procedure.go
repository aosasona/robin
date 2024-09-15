package types

import (
	"encoding/json"
	"reflect"
)

type ProcedureType string

const (
	ProcedureTypeQuery    ProcedureType = "query"
	ProcedureTypeMutation ProcedureType = "mutation"
)

type Procedure interface {
	// The name of the procedure
	Name() string

	// The type of the procedure, one of 'query' or 'mutation'
	Type() ProcedureType

	// Return an empty type that represents the payload that the procedure expects
	// WARNING: whatever is returned here is only used for type inference/reflection during runtime; no value should be expected here
	PayloadInterface() any

	// Return an empty type that represents the return value of the procedure
	// WARNING: whatever is returned here is only used for type inference/reflection during runtime; no value should be expected here
	ReturnInterface() any

	// Check if the procedure expects a payload or not
	// This is useful for procedures that don't expect a payload, so we can instantly skip the payload decoding step
	ExpectsPayload() bool

	// Call the procedure with the given context and payload
	Call(*Context, any) (any, error)
}

type JsonSerializable interface {
	json.Marshaler
	json.Unmarshaler
}

// No-op type to represent a procedure that doesn't return any response or take any payload
type Void struct{}

func (v Void) MarshalJSON() ([]byte, error) {
	return []byte("null"), nil
}

func (v *Void) UnmarshalJSON(data []byte) error {
	if string(data) != "" && string(data) != "null" {
		return &json.UnsupportedValueError{
			Value: reflect.ValueOf(data),
			Str:   string(data),
		}
	}
	return nil
}

var _ JsonSerializable = (*Void)(nil)
