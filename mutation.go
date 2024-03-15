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

func (q *mutation[_, _]) Type() ProcedureType {
	return MutationProcedure
}

func Mutation[R any, B any](name string, fn MutationFn[R, B]) *mutation[R, B] {
	return &mutation[R, B]{name: name, fn: fn}
}
