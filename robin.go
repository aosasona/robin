package robin

import (
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
)

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
	PayloadInterface() any
	Call(*Context, any) (any, error)
	StripIllegalChars()
}

type (
	Robin struct {
		// Enable debug mode to log useful info
		debug bool

		// a list of query and mutation procedures
		procedures map[string]Procedure

		// a function that will be called when an error occurs, if not provided, the default error handler will be used
		errorHandler ErrorHandler
	}

	Context struct {
		// The raw incoming request
		Request *http.Request

		// The raw response writer
		Response http.ResponseWriter

		// The name of the procedure
		ProcedureName string

		// The type of the procedure
		ProcedureType ProcedureType
	}
)

type Options struct {
	EnableDebugMode bool

	// ErrorHandler is a function that will be called when an error occurs, it should ideally return a marshallable struct
	ErrorHandler ErrorHandler
}

func DefaultErrorHandler(err error) ([]byte, int) {
	var (
		code    int    = 500
		message string = err.Error()
	)

	if e, ok := err.(Error); ok {
		message = e.Message

		if e.Code >= 400 && e.Code < 600 {
			code = e.Code
		}
	} else if e, ok := err.(InternalError); ok {
		message = e.Reason

		slog.Error("An internal error occurred", slog.String("reason", e.Reason), slog.Any("originalError", e.OriginalError.Error()))
	}

	return []byte(message), code
}

// Robin is just going to be an adapter for something like Echo
func New(opts *Options) *Robin {
	errorHandler := DefaultErrorHandler
	if opts.ErrorHandler != nil {
		errorHandler = opts.ErrorHandler
	}

	return &Robin{
		debug:        opts.EnableDebugMode,
		procedures:   make(map[string]Procedure),
		errorHandler: errorHandler,
	}
}

func (r *Robin) Add(procedure Procedure) *Robin {
	procedure.StripIllegalChars()

	if _, ok := r.procedures[procedure.Name()]; ok {
		if r.debug {
			slog.Warn("Attempted to add a duplicate procedure, skipping...", slog.String("procedureName", procedure.Name()))
		}

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
	defer func(r *Robin) {
		if e := recover(); e != nil {
			r.sendError(w, InternalError{Reason: fmt.Sprintf("Panic trapped: %v", e)})
		}
	}(r)

	defer req.Body.Close()

	var err error

	ctx := &Context{Request: req, Response: w}

	ctx.ProcedureType, ctx.ProcedureName, err = r.getProcedureMetaFromURL(req.URL)
	if err != nil {
		r.sendError(w, err)
		return
	}

	procedure, found := r.findProcedure(ctx.ProcedureName, ctx.ProcedureType)
	if !found {
		r.sendError(w, InternalError{Reason: "Procedure not found"})
		return
	}

	switch ProcedureType(ctx.ProcedureType) {
	case ProcedureTypeQuery, ProcedureTypeMutation:
		err := r.handleProcedureCall(ctx, *procedure)
		if err != nil {
			r.sendError(w, err)
			return
		}
	default:
		r.sendError(w, InternalError{Reason: "Invalid procedure type, expect one of 'query' or 'mutation', got " + string(ctx.ProcedureType)})
		return
	}
}

func (r *Robin) sendError(w http.ResponseWriter, err error) {
	if r.debug {
		slog.Error("An error occurred in handler", slog.Any("error", err))
	}

	errorResp, code := r.errorHandler(err)
	jsonResp := fmt.Sprintf(`{"error": "%s"}`, string(errorResp))

	w.Header().Add("content-type", "application/json")
	w.WriteHeader(code)
	w.Write([]byte(jsonResp))
}
