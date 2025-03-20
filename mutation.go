package robin

import (
	"fmt"
	"reflect"
	"strings"

	"go.trulyao.dev/robin/internal/guarded"
	"go.trulyao.dev/robin/types"
)

type mutation[ReturnType any, BodyType any] struct {
	*baseProcedure[ReturnType, BodyType]
}

// Creates a new mutation with the given name and handler function
func Mutation[R any, B any](name string, fn ProcedureFn[R, B]) *mutation[R, B] {
	var body B
	expectedPayloadType := guarded.ExpectsPayload(body)

	m := &mutation[R, B]{
		baseProcedure: &baseProcedure[R, B]{
			name:                name,
			fn:                  fn,
			expectedPayloadType: expectedPayloadType,
			excludedMiddleware:  &types.ExclusionList{},
		},
	}
	m.alias = m.NormalizeProcedureName()

	return m
}

// Alias for `Mutation` to create a new mutation procedure
func M[R any, B any](name string, fn ProcedureFn[R, B]) *mutation[R, B] {
	return Mutation(name, fn)
}

// Creates a new mutation with the given name, handler function, and middleware functions
func MutationWithMiddleware[R any, B any](
	name string,
	fn ProcedureFn[R, B],
	middleware ...types.Middleware,
) *mutation[R, B] {
	m := Mutation(name, fn)
	m.WithMiddleware(middleware...)
	return m
}

// Returns the type of the procedure, one of 'query' or 'mutation' - in this case, it's always 'mutation'
func (m *mutation[_, _]) Type() ProcedureType {
	return ProcedureTypeMutation
}

// String returns a string representation of the mutation
func (m *mutation[_, _]) String() string {
	return fmt.Sprintf("Mutation(%s)", m.name)
}

// NormalizeProcedureName normalizes the procedure name to a more human-readable format for use in the REST API
func (m *mutation[_, _]) NormalizeProcedureName() string {
	var alias string

	// Replace all non-alphanumeric characters with dot
	alias = ReAlphaNumeric.ReplaceAllString(m.name, ".")

	// Replace all multiple dots with a single dot
	alias = ReIllegalDot.ReplaceAllString(alias, ".")

	// Remove all words that are associable with the query type
	alias = ReMutationWords.ReplaceAllString(alias, "")

	// Remove all leading and trailing dots and spaces
	alias = strings.TrimSpace(alias)
	alias = strings.Trim(alias, ".")

	return alias
}

// WithAlias sets the alias of the query
func (m *mutation[_, _]) WithAlias(alias string) Procedure {
	m.alias = alias
	return m
}

// Calls the mutation with the given context and body
func (m *mutation[ReturnType, BodyType]) Call(ctx *Context, rawBody any) (any, error) {
	body, err := guarded.CastType(rawBody, m.in.InferredType())
	if err != nil {
		return nil, err
	}

	if m.fn == nil {
		return nil, RobinError{Reason: fmt.Sprintf("Procedure %s has no function attached", m.name)}
	}

	return m.fn(ctx, body)
}

// Validate validates the query
func (m *mutation[_, _]) Validate() error {
	// Check if the query name is valid
	if m.name == "" {
		return RobinError{Reason: "Query name cannot be empty"}
	}

	if !ReValidProcedureName.MatchString(m.name) {
		return RobinError{
			Reason: fmt.Sprintf(
				"Invalid procedure name: `%s`, expected string matching regex `%s` (example: `get_user`, `todo.create`)",
				m.name,
				ReValidProcedureName,
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
