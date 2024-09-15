package robin

import (
	"fmt"

	"go.trulyao.dev/robin/internal/guarded"
)

type (
	QueryFn[ReturnType any, ParamsType any] func(ctx *Context, body ParamsType) (ReturnType, error)

	query[ReturnType any, ParamsType any] struct {
		// The name of the query
		name string

		// The function that will be called when the query is executed
		fn QueryFn[ReturnType, ParamsType]

		// A placeholder for the type of the body that the query expects
		// WARNING: This never really has a value, it's just used for "type inference/reflection" during runtime
		in ParamsType

		// A placeholder for the return type of the query
		// WARNING: This never really has a value, it's just used for "type inference/reflection" during runtime
		out ReturnType

		// Whether the query expects a payload or not
		expectsPayload bool
	}
)

// Name returns the name of the query
func (q *query[_, _]) Name() string {
	return q.name
}

// Returns the type of the procedure, one of 'query' or 'mutation' - in this case, it's always 'query'
func (q *query[_, _]) Type() ProcedureType {
	return ProcedureTypeQuery
}

// PayloadInterface returns a placeholder variable with the type of the payload that the query expects, this value is empty and only used for type inference/reflection during runtime
func (q *query[_, _]) PayloadInterface() any {
	return q.in
}

// ReturnInterface returns a placeholder variable with the type of the return value of the query, this value is empty and only used for type inference/reflection during runtime
func (q *query[_, _]) ReturnInterface() any {
	return q.out
}

// Calls the query with the given context and params
func (q *query[ReturnType, ParamsType]) Call(ctx *Context, rawParams any) (any, error) {
	params, err := guarded.CastType(rawParams, q.in)
	if err != nil {
		return nil, err
	}

	if q.fn == nil {
		return nil, RobinError{Reason: fmt.Sprintf("Procedure %s has no function attached", q.name)}
	}

	return q.fn(ctx, params)
}

// ExpectsPayload returns whether the query expects a payload or not
func (q *query[_, _]) ExpectsPayload() bool {
	return q.expectsPayload
}

// Creates a new query with the given name and handler function
func Query[R any, B any](name string, fn QueryFn[R, B]) *query[R, B] {
	name = string(procedureNameRegex.ReplaceAll([]byte(name), []byte("")))

	var body B
	expectsPayload := guarded.ExpectsPayload(body)

	return &query[R, B]{name: name, fn: fn, expectsPayload: expectsPayload}
}

var _ Procedure = (*query[any, any])(nil)
