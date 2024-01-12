package robin

import "net/http"

type ProcedureType int

const (
	QueryProcedure ProcedureType = iota
	MutationProcedure
)

type Procedure interface {
	Name() string
	Type() ProcedureType
}

type ErrorHandler func(error) any

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

func (r *Robin) Add(procedure Procedure) *Robin {
	r.procedures = append(r.procedures, procedure)
	return r
}

func (r *Robin) AddProcedure(procedure Procedure) *Robin {
	return r.Add(procedure)
}
