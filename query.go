package robin

import (
	"fmt"
)

type (
	QueryFn[ReturnType any, ParamsType any] func(ctx *Context, body ParamsType) (ReturnType, error)

	query[ReturnType any, ParamsType any] struct {
		name   string
		fn     QueryFn[ReturnType, ParamsType]
		params ParamsType // The type of the params that the query expects, this never really has a value, it's just used for "type checking" during runtime
	}
)

func (q *query[_, _]) Name() string {
	return q.name
}

func (q *query[_, _]) Type() ProcedureType {
	return ProcedureTypeQuery
}

func (q *query[ReturnType, ParamsType]) PayloadInterface() any {
	return q.params
}

func (q *query[ReturnType, ParamsType]) Call(ctx *Context, rawParams any) (any, error) {
	params, err := guardedCast(rawParams, q.params)
	if err != nil {
		return nil, err
	}

	if q.fn == nil {
		return nil, RobinError{Reason: fmt.Sprintf("Procedure %s has no function attached", q.name)}
	}

	return q.fn(ctx, params)
}

func (q *query[_, _]) StripIllegalChars() {
	procedureNameRegex.ReplaceAll([]byte(q.name), []byte(""))
}

func Query[R any, B any](name string, fn QueryFn[R, B]) *query[R, B] {
	return &query[R, B]{name: name, fn: fn}
}

var _ Procedure = &query[any, any]{}
