package robin

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

func (m *mutation[_, _]) StripIllegalChars() {
	mutationNameRegex.ReplaceAll([]byte(m.name), []byte(""))
}

func Mutation[R any, B any](name string, fn MutationFn[R, B]) *mutation[R, B] {
	return &mutation[R, B]{name: name, fn: fn}
}

var _ Procedure = &mutation[string, string]{}
