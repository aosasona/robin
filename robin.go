package robin

import "net/http"

type ProdecureType int

const (
	QueryProcedure ProdecureType = iota
	MutationProcedure
)

type Procedure interface {
	Name() string
	Type() ProdecureType
	Body() any
}

type (
	ProcedureFn[ReturnType any] func(ctx Context) (ReturnType, error)
	ErrorHandler                func(error) interface{}
)

type query[ReturnType any] struct {
	name string
	fn   ProcedureFn[ReturnType]
}

type Robin struct {
	procedures   []Procedure
	errorHandler ErrorHandler
}

type Context struct {
	Request  *http.Request
	Response http.ResponseWriter
}

// Robin is just going to be an adapter for something like Echo
func New(errorHandler ErrorHandler) *Robin {
	return &Robin{procedures: []Procedure{}, errorHandler: errorHandler}
}

func Query[R any](name string, fn ProcedureFn[R]) *query[R] {
	return &query[R]{name: name, fn: fn}
}
