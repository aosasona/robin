package robin

import (
	"fmt"
	"strings"

	"go.trulyao.dev/robin/internal/guarded"
	"go.trulyao.dev/robin/types"
)

type query[ReturnType any, ParamsType any] struct {
	*baseProcedure[ReturnType, ParamsType]
}

// Creates a new query with the given name and handler function
func Query[R any, B any](name string, fn ProcedureFn[R, B]) *query[R, B] {
	var body B
	expectedPayloadType := guarded.ExpectsPayload(body)

	q := &query[R, B]{
		baseProcedure: &baseProcedure[R, B]{
			name:                name,
			fn:                  fn,
			expectedPayloadType: expectedPayloadType,
			excludedMiddleware:  &types.ExclusionList{},
		},
	}
	q.alias = q.NormalizeProcedureName()

	return q
}

// Alias for `Query` to create a new query procedure
func Q[R any, B any](name string, fn ProcedureFn[R, B]) *query[R, B] {
	return Query(name, fn)
}

// Creates a new query with the given name, handler function and middleware functions
func QueryWithMiddleware[R any, B any](
	name string,
	fn ProcedureFn[R, B],
	middleware ...types.Middleware,
) *query[R, B] {
	q := Query(name, fn)
	q.middlewareFns = middleware
	return q
}

// Returns the type of the procedure, one of 'query' or 'mutation' - in this case, it's always 'query'
func (q *query[_, _]) Type() ProcedureType {
	return ProcedureTypeQuery
}

// String returns a string representation of the query
func (q *query[_, _]) String() string {
	return fmt.Sprintf("Query(%s)", q.name)
}

// NormalizeProcedureName normalizes the procedure name to a more human-readable format for use in the REST API
func (q *query[_, _]) NormalizeProcedureName() string {
	var alias string

	// Replace all non-alphanumeric characters with dot
	alias = ReAlphaNumeric.ReplaceAllString(q.name, ".")

	// Replace all multiple dots with a single dot
	alias = ReIllegalDot.ReplaceAllString(alias, ".")

	// Remove all words that are associable with the query type
	alias = ReQueryWords.ReplaceAllString(alias, "")

	// Remove all leading and trailing dots and spaces
	alias = strings.TrimSpace(alias)
	alias = strings.Trim(alias, ".")

	return alias
}

// WithAlias sets the alias of the query
func (q *query[_, _]) WithAlias(alias string) Procedure {
	q.alias = alias
	return q
}

// Calls the query with the given context and params
func (q *query[ReturnType, ParamsType]) Call(ctx *Context, rawParams any) (any, error) {
	params, err := guarded.CastType(rawParams, q.in.InferredType())
	if err != nil {
		return nil, err
	}

	if q.fn == nil {
		return nil, RobinError{Reason: fmt.Sprintf("Procedure %s has no function attached", q.name)}
	}

	return q.fn(ctx, params)
}

// ExpectsPayload returns whether the query expects a payload or not
func (q *query[_, _]) ExpectedPayloadType() types.ExpectedPayloadType {
	return q.expectedPayloadType
}

// Validate validates the query
func (q *query[_, _]) Validate() error {
	// Check if the query name is valid
	if q.name == "" {
		return RobinError{Reason: "Query name cannot be empty"}
	}

	if !ReValidProcedureName.MatchString(q.name) {
		return RobinError{
			Reason: fmt.Sprintf(
				"Invalid procedure name: `%s`, expected string matching regex `%s` (example: `get_user`, `todo.create`)",
				q.name,
				ReValidProcedureName,
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

// ExcludeMiddleware takes a list of global middleware names and excludes them from the query
func (q *query[_, _]) ExcludeMiddleware(names ...string) types.Procedure {
	q.excludedMiddleware.AddMany(names)
	return q
}

// ExcludedMiddleware returns the list of middleware functions that are excluded from the query
func (q *query[_, _]) ExcludedMiddleware() *types.ExclusionList {
	return q.excludedMiddleware
}

// WithRawPayload sets the type of the payload that the query expects (for client type inference)
func (q *query[_, _]) WithRawPayload(actualPayloadType any) Procedure {
	// Ensure that the original input (provided via generics in In) is io.ReadCloser
	mustImplementReadCloser(q.fn, ProcedureTypeQuery)

	q.in.SetOverrideType(actualPayloadType)
	q.expectedPayloadType = types.ExpectedPayloadRaw
	return q
}

var _ Procedure = (*query[any, any])(nil)
