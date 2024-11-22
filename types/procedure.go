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

type JSONSerializable interface {
	json.Marshaler
	json.Unmarshaler
}

type HttpMethod string

const (
	HttpMethodGet    HttpMethod = "GET"
	HttpMethodPost   HttpMethod = "POST"
	HttpMethodPut    HttpMethod = "PUT"
	HttpMethodPatch  HttpMethod = "PATCH"
	HttpMethodDelete HttpMethod = "DELETE"
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

	// Validate the procedure
	Validate() error

	// Middleware to be executed before the procedure is called
	MiddlewareFns() []Middleware

	// Set the middleware functions for the procedure at the beginning of the middleware chain
	// You ideally should not use this method, use WithMiddleware instead unless you absolutely need to prepend middleware functions to the chain
	PrependMiddleware(...Middleware) Procedure

	// Set the middleware functions for the procedure
	WithMiddleware(...Middleware) Procedure

	ExcludedMiddleware() *ExclusionList

	// Exclude middleware functions from the procedure
	ExcludeMiddleware(...string) Procedure

	// Alias the procedure with a different name for the REST API (and other potential future use cases)
	WithAlias(string) Procedure

	// Get the alias of the Procedure if it has one, otherwise, it returns a normalized version of the procedure name (e.g. `get_user` -> `user`, `todo.create` -> `todo.create`, `find-users` -> `users`)
	//
	// Common words like `get`, `find`, `create`, `update`, `delete` are normalized to their respective actions based on the procedure type
	Alias() string
}

// No-op type to represent a procedure that doesn't return any response or take any payload
type (
	_RobinVoid struct{} // Used for identification of robin's special void type

	Void = _RobinVoid
)

func (v _RobinVoid) MarshalJSON() ([]byte, error) {
	return []byte("null"), nil
}

func (v *_RobinVoid) UnmarshalJSON(data []byte) error {
	if string(data) != "" && string(data) != "null" {
		return &json.UnsupportedValueError{
			Value: reflect.ValueOf(data),
			Str:   string(data),
		}
	}
	return nil
}

var _ JSONSerializable = (*Void)(nil)
