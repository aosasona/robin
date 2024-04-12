package robin

import "fmt"

type (
	MutationFn[ReturnType any, BodyType any] func(ctx *Context, body BodyType) (ReturnType, error)

	mutation[ReturnType any, BodyType any] struct {
		name string
		body BodyType
		fn   MutationFn[ReturnType, BodyType]
	}
)

func (m *mutation[_, _]) Name() string {
	return m.name
}

func (m *mutation[_, _]) Type() ProcedureType {
	return ProcedureTypeMutation
}

func (m *mutation[_, _]) PayloadInterface() any {
	return m.body
}

func (m *mutation[ReturnType, BodyType]) Call(ctx *Context, rawBody any) (any, error) {
	body, err := guardedCast(rawBody, m.body)
	if err != nil {
		return nil, err
	}

	if m.fn == nil {
		return nil, RobinError{Reason: fmt.Sprintf("Procedure %s has no function attached", m.name)}
	}

	return m.fn(ctx, body)
}

func (m *mutation[_, _]) StripIllegalChars() {
	procedureNameRegex.ReplaceAll([]byte(m.name), []byte(""))
}

func Mutation[R any, B any](name string, fn MutationFn[R, B]) *mutation[R, B] {
	return &mutation[R, B]{name: name, fn: fn}
}

var _ Procedure = &mutation[any, any]{}
