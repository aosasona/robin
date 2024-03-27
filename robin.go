package robin

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// TODO: use a single query param to decide what the procedure and type is e.g. q__[name]
const (
	ProcSeparator = "__"
	ProcNameKey   = ProcSeparator + "proc"
)

var procedureNameRegex = regexp.MustCompile(`(?m)[^a-zA-Z0-9]`)

type ProcedureType string

const (
	ProcedureTypeQuery    ProcedureType = "query"
	ProcedureTypeMutation ProcedureType = "mutation"
)

type Procedure interface {
	Name() string
	Type() ProcedureType
	StripIllegalChars()
}

type ErrorHandler func(error) (any, int)

type Robin struct {
	// a list of query and mutation procedures
	procedures map[string]Procedure

	// a function that will be called when an error occurs, if not provided, the default error handler will be used
	errorHandler ErrorHandler
}

type Context struct {
	Request  *http.Request
	Response http.ResponseWriter

	// TODO: add fields for extracting body, query, etc
}

type Options struct {
	// ErrorHandler is a function that will be called when an error occurs, it should ideally return a marshallable struct
	ErrorHandler ErrorHandler
}

func DefaultErrorHandler(err error) (any, int) {
	return struct{ message string }{message: err.Error()}, 500
}

// Robin is just going to be an adapter for something like Echo
func New(opts *Options) *Robin {
	errorHandler := DefaultErrorHandler
	if opts.ErrorHandler != nil {
		errorHandler = opts.ErrorHandler
	}

	return &Robin{
		procedures:   make(map[string]Procedure),
		errorHandler: errorHandler,
	}
}

func (r *Robin) Add(procedure Procedure) *Robin {
	procedure.StripIllegalChars()

	if _, ok := r.procedures[procedure.Name()]; ok {
		return r
	}

	r.procedures[procedure.Name()] = procedure
	return r
}

func (r *Robin) AddProcedure(procedure Procedure) *Robin {
	return r.Add(procedure)
}

func (r *Robin) Handler() http.HandlerFunc {
	return r.ServeHTTP
}

func (r *Robin) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := &Context{Request: req, Response: w}

	procedureType, procedureName, err := r.getProcedureMetaFromURL(req.URL)
	if err != nil {
		r.sendError(w, err)
		return
	}

	switch ProcedureType(procedureType) {
	case ProcedureTypeQuery:
		r.handleQuery(ctx, "")
	case ProcedureTypeMutation:
		r.handleMutation(ctx, "")
	default:
	}
}

func (r *Robin) handleQuery(ctx *Context, name string) (any, error) {
	return nil, nil
}

func (r *Robin) handleMutation(ctx *Context, name string) (any, error) {
	return nil, nil
}

func (r *Robin) getProcedureMetaFromURL(url *url.URL) (ProcedureType, string, error) {
	var (
		procedureName string
		procedureType ProcedureType
	)

	// Queries can only be issued via GET requests, and Mutations can only be issued via POST requests but in both cases, the procedure name is attached to the URL query
	proc := url.Query().Get(ProcNameKey)
	if strings.TrimSpace(procedureName) == "" {
		return "", "", errors.New("No procedure name provided")
	}

	procParts := strings.Split(proc, ProcSeparator)
	if len(procParts) != 2 {
		return "", "", fmt.Errorf("Invalid procedure param provided in URL, expected format (q|m)%s[name] e.g q%sgetUser", ProcSeparator, ProcSeparator)
	}

	shortProcType := procParts[0]
	switch shortProcType {
	case "q":
		procedureType = ProcedureTypeQuery
	case "m":
		procedureType = ProcedureTypeMutation
	default:
		return "", "", errors.New("No procedure name provided")
	}

	return procedureType, procedureName, nil
}

func (r *Robin) findProcedure(name string, procedureType ProcedureType) *Procedure, bool {
	return nil, nil
}

func (r *Robin) sendError(w http.ResponseWriter, err error) {
	errorResp, code := r.errorHandler(err)
	jsonResp := toJsonBytes(errorResp)

	w.WriteHeader(code)
	w.Header().Add("Content-Type", "application/json")
	w.Write(jsonResp)
}

func toJsonBytes(data any) []byte {
	var jsonData []byte = []byte{}

	return jsonData
}
