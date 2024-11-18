package robin

import (
	"fmt"

	"go.trulyao.dev/robin/internal/guarded"
	"go.trulyao.dev/robin/types"
)

type (
	MutationFn[ReturnType any, BodyType any] func(ctx *Context, body BodyType) (ReturnType, error)

	mutation[ReturnType any, BodyType any] struct {
		// The name of the mutation
		name string

		// The function that will be called when the mutation is executed
		fn MutationFn[ReturnType, BodyType]

		// A placeholder for the type of the body that the mutation expects
		// WARNING: This never really has a value, it's just used for "type inference/reflection" during runtime
		in BodyType

		// A placeholder for the return type of the mutation
		// WARNING: This never really has a value, it's just used for "type inference/reflection" during runtime
		out ReturnType

		// Middleware functions to be executed before the mutation is called
		middlewareFns []types.Middleware

		// Whether the mutation expects a payload or not
		expectsPayload bool

		// Excluded middleware functions
		excludedMiddleware *types.ExclusionList
	}
)

// Returns the name of the current mutation
func (m *mutation[_, _]) Name() string {
	return m.name
}

// Returns the type of the procedure, one of 'query' or 'mutation' - in this case, it's always 'mutation'
func (m *mutation[_, _]) Type() ProcedureType {
	return ProcedureTypeMutation
}

// PayloadInterface returns a placeholder variable with the type of the payload that the mutation expects, this value is empty and only used for type inference/reflection during runtime
func (m *mutation[_, _]) PayloadInterface() any {
	return m.in
}

// ReturnInterface returns a placeholder variable with the type of the return value of the mutation, this value is empty and only used for type inference/reflection during runtime
func (m *mutation[_, _]) ReturnInterface() any {
	return m.out
}

// Calls the mutation with the given context and body
func (m *mutation[ReturnType, BodyType]) Call(ctx *Context, rawBody any) (any, error) {
	body, err := guarded.CastType(rawBody, m.in)
	if err != nil {
		return nil, err
	}

	if m.fn == nil {
		return nil, RobinError{Reason: fmt.Sprintf("Procedure %s has no function attached", m.name)}
	}

	return m.fn(ctx, body)
}

// Returns whether the mutation expects a payload or not
func (m *mutation[_, _]) ExpectsPayload() bool {
	return m.expectsPayload
}

// Validate validates the query
func (m *mutation[_, _]) Validate() error {
	// Check if the query name is valid
	if m.name == "" {
		return RobinError{Reason: "Query name cannot be empty"}
	}

	if !procedureNameRegex.MatchString(m.name) {
		return RobinError{
			Reason: fmt.Sprintf(
				"Invalid procedure name: `%s`, expected string matching regex `%s` (example: `get_user`, `todo.create`)",
				m.name,
				procedureNameRegex,
			),
		}
	}

	return nil
}

// MiddlewareFns returns the middleware functions that should be executed before the mutation is called
func (m *mutation[_, _]) MiddlewareFns() []types.Middleware {
	return m.middlewareFns
}

// PrependMiddleware sets the middleware functions for the mutation at the beginning of the middleware chain
func (m *mutation[_, _]) PrependMiddleware(fns ...types.Middleware) Procedure {
	m.middlewareFns = append(fns, m.middlewareFns...)
	return m
}

// WithMiddleware sets the middleware functions for the mutation
func (m *mutation[_, _]) WithMiddleware(fns ...types.Middleware) Procedure {
	m.middlewareFns = append(m.middlewareFns, fns...)
	return m
}

// Creates a new mutation with the given name and handler function
func Mutation[R any, B any](name string, fn MutationFn[R, B]) *mutation[R, B] {
	var body B
	expectsPayload := guarded.ExpectsPayload(body)

	return &mutation[R, B]{name: name, fn: fn, expectsPayload: expectsPayload, excludedMiddleware: &types.ExclusionList{}}
}

// Alias for `Mutation` to create a new mutation procedure
func M[R any, B any](name string, fn MutationFn[R, B]) *mutation[R, B] {
	return Mutation(name, fn)
}

// Creates a new mutation with the given name, handler function, and middleware functions
func MutationWithMiddleware[R any, B any](
	name string,
	fn MutationFn[R, B],
	middleware ...types.Middleware,
) *mutation[R, B] {
	m := Mutation(name, fn)
	m.WithMiddleware(middleware...)
	return m
}

// ExcludeMiddleware takes a list of global middleware names and excludes them from the mutation
func (m *mutation[_, _]) ExcludeMiddleware(names ...string) types.Procedure {
	m.excludedMiddleware.AddMany(names)
	return m
}

// ExcludedMiddleware returns the list of middleware functions that are excluded from the query
func (m *mutation[_, _]) ExcludedMiddleware() *types.ExclusionList {
	return m.excludedMiddleware
}

var _ Procedure = (*mutation[any, any])(nil)
