package robin

import (
	"fmt"

	"go.trulyao.dev/robin/internal/guarded"
)

type (
	MutationFn[ReturnType any, BodyType any] func(ctx *Context, body BodyType) (ReturnType, error)

	mutation[ReturnType any, BodyType any] struct {
		// The name of the mutation
		name string

		// The function that will be called when the mutation is executed
		fn MutationFn[ReturnType, BodyType]

		// The type of the body that the mutation expects
		// WARNING: This never really has a value, it's just used for "type inference/reflection" during runtime
		body BodyType

		// Whether the mutation expects a payload or not
		expectsPayload bool
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

// Returns the type of the payload that the mutation expects, this value is empty and only used for type inference/reflection during runtime
func (m *mutation[_, _]) PayloadInterface() any {
	return m.body
}

// Calls the mutation with the given context and body
func (m *mutation[ReturnType, BodyType]) Call(ctx *Context, rawBody any) (any, error) {
	body, err := guarded.CastType(rawBody, m.body)
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

// Creates a new mutation with the given name and handler function
func Mutation[R any, B any](name string, fn MutationFn[R, B]) *mutation[R, B] {
	name = string(procedureNameRegex.ReplaceAll([]byte(name), []byte("")))

	var body B
	expectsPayload := guarded.ExpectsPayload(body)

	return &mutation[R, B]{name: name, fn: fn, expectsPayload: expectsPayload}
}

var _ Procedure = (*mutation[any, any])(nil)
