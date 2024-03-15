package robin

import (
	"fmt"
	"net/http"
)

type ProcedureType string

const (
	QueryProcedure    ProcedureType = "query"
	MutationProcedure ProcedureType = "mutation"
)

type Procedure interface {
	Name() string
	Type() ProcedureType
}

type ErrorHandler func(error) any

type Robin struct {
	// a list of query and mutation procedures
	procedures []Procedure

	// a function that will be called when an error occurs, if not provided, the default error handler will be used
	errorHandler ErrorHandler
}

type Context struct {
	Request  *http.Request
	Response http.ResponseWriter

	// TODO: add fields for extracting body, query, etc
}

type Options struct {
	// ErrorHandler is a function that will be called when an error occurs
	ErrorHandler ErrorHandler
}

// Robin is just going to be an adapter for something like Echo
func New(opts *Options) *Robin {
	return &Robin{
		procedures:   []Procedure{},
		errorHandler: opts.ErrorHandler,
	}
}

func (r *Robin) Add(procedure Procedure) *Robin {
	r.procedures = append(r.procedures, procedure)
	return r
}

func (r *Robin) AddProcedure(procedure Procedure) *Robin {
	return r.Add(procedure)
}

func (r *Robin) Handler() http.HandlerFunc {
	return r.ServeHTTP
}

func (r *Robin) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	_ = &Context{Request: req, Response: w}

	fmt.Println("Hello, world!")
}

func (r *Robin) handleQuery(ctx *Context, name string, body any) (any, error) {
	return nil, nil
}

func (r *Robin) handleMutation(ctx *Context, name string, body any) (any, error) {
	return nil, nil
}
