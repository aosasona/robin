package robin

type (
	QueryFn[ReturnType any, ParamsType any] func(ctx *Context, body ParamsType) (ReturnType, error)

	query[ReturnType any, ParamsType any] struct {
		name   string
		fn     QueryFn[ReturnType, ParamsType]
		params ParamsType
	}
)

func (q *query[_, _]) Name() string {
	return q.name
}

func (q *query[_, _]) Type() ProcedureType {
	return ProcedureTypeQuery
}

func (q *query[_, _]) StripIllegalChars() {
	compressedProcedureNameRegex.ReplaceAll([]byte(q.name), []byte(""))
}

func Query[R any, B any](name string, fn QueryFn[R, B]) *query[R, B] {
	return &query[R, B]{name: name, fn: fn}
}

var _ Procedure = &query[string, string]{}
