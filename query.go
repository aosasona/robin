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

		// The type of the params that the query expects
		// WARNING: This never really has a value, it's just used for "type inference/reflection" during runtime
		params ParamsType
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

// PayloadInterface returns the type of the payload that the query expects, this value is empty and only used for type inference/reflection during runtime
func (q *query[ReturnType, ParamsType]) PayloadInterface() any {
	return q.params
}

// Calls the query with the given context and params
func (q *query[ReturnType, ParamsType]) Call(ctx *Context, rawParams any) (any, error) {
	params, err := guarded.CastType(rawParams, q.params)
	if err != nil {
		return nil, err
	}

	if q.fn == nil {
		return nil, RobinError{Reason: fmt.Sprintf("Procedure %s has no function attached", q.name)}
	}

	return q.fn(ctx, params)
}

// Creates a new query with the given name and handler function
func Query[R any, B any](name string, fn QueryFn[R, B]) *query[R, B] {
	name = string(procedureNameRegex.ReplaceAll([]byte(name), []byte("")))
	return &query[R, B]{name: name, fn: fn}
}

var _ Procedure = (*query[any, any])(nil)
