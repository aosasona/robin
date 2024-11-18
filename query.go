package robin

import (
	"fmt"

	"go.trulyao.dev/robin/internal/guarded"
	"go.trulyao.dev/robin/types"
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

		// Middleware functions to be executed before the mutation is called
		middlewareFns []types.Middleware

		// Whether the query expects a payload or not
		expectsPayload bool

		// Excluded middleware functions
		excludedMiddleware *types.ExclusionList
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

// Validate validates the query
func (q *query[_, _]) Validate() error {
	// Check if the query name is valid
	if q.name == "" {
		return RobinError{Reason: "Query name cannot be empty"}
	}

	if !procedureNameRegex.MatchString(q.name) {
		return RobinError{
			Reason: fmt.Sprintf(
				"Invalid procedure name: `%s`, expected string matching regex `%s` (example: `get_user`, `todo.create`)",
				q.name,
				procedureNameRegex,
			),
		}
	}

	return nil
}

// MiddlewareFns returns the middleware functions to be executed before the query is called
func (q *query[_, _]) MiddlewareFns() []types.Middleware {
	return q.middlewareFns
}

// PrependMiddleware sets the middleware functions for the mutation at the beginning of the middleware chain
func (q *query[_, _]) PrependMiddleware(fns ...types.Middleware) Procedure {
	q.middlewareFns = append(fns, q.middlewareFns...)
	return q
}

// Add the middleware functions for the query
func (q *query[_, _]) WithMiddleware(fns ...types.Middleware) Procedure {
	q.middlewareFns = append(q.middlewareFns, fns...)
	return q
}

// Creates a new query with the given name and handler function
func Query[R any, B any](name string, fn QueryFn[R, B]) *query[R, B] {
	var body B
	expectsPayload := guarded.ExpectsPayload(body)

	return &query[R, B]{name: name, fn: fn, expectsPayload: expectsPayload, excludedMiddleware: &types.ExclusionList{}}
}

// Alias for `Query` to create a new query procedure
func Q[R any, B any](name string, fn QueryFn[R, B]) *query[R, B] {
	return Query(name, fn)
}

// Creates a new query with the given name, handler function and middleware functions
func QueryWithMiddleware[R any, B any](
	name string,
	fn QueryFn[R, B],
	middleware ...types.Middleware,
) *query[R, B] {
	q := Query(name, fn)
	q.middlewareFns = middleware
	return q
}

// ExcludeMiddleware takes a list of global middleware names and excludes them from the query
func (q *query[_, _]) ExcludeMiddleware(names ...string) types.Procedure {
	q.excludedMiddleware.AddMany(names)
	return q
}

// ExcludedMiddleware returns the list of middleware functions that are excluded from the query
func (q *query[_, _]) ExcludedMiddleware() *types.ExclusionList {
	return q.excludedMiddleware
}

var _ Procedure = (*query[any, any])(nil)
