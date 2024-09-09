package robin

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strings"
)

// TODO: add robin.Void type to represent a procedure that doesn't return any response or take any payload

const (
	ProcSeparator = "__"
	ProcNameKey   = ProcSeparator + "proc"

	EnvRobinEnableTSGen = "ROBIN_ENABLE_SCHEMA_GEN"
)

var procedureNameRegex = regexp.MustCompile(`(?m)[^a-zA-Z0-9]`)

type ProcedureType string

const (
	ProcedureTypeQuery    ProcedureType = "query"
	ProcedureTypeMutation ProcedureType = "mutation"
)

type Procedure interface {
	// The name of the procedure
	Name() string

	// The type of the procedure, one of 'query' or 'mutation'
	Type() ProcedureType

	// Return an empty type that represents the payload that the procedure expects
	// WARNING: whatever is returned here is only used for type inference/reflection during runtime; no value should be expected here
	PayloadInterface() any

	// Call the procedure with the given context and payload
	Call(*Context, any) (any, error)
}

type (
	Robin struct {
		// Path to the generated typescript schema
		bindingsPath string

		// Enable the cache for the resolved procedure types
		cacheResolvedTypes bool

		// Enable the generation of typescript schema during runtime, this is disabled by default to prevent unnecessary overhead when not needed
		enableTypescriptGen bool

		// Enable debug mode to log useful info
		debug bool

		// A list of query and mutation procedures
		procedures map[string]Procedure

		// A function that will be called when an error occurs, if not provided, the default error handler will be used
		errorHandler ErrorHandler

		// TODO: add resolved types cache
	}
)

type Options struct {
	// Path to the generated typescript schema
	BindingPath string

	// Enable the cache for the resolved procedure types
	CacheResolvedTypes bool

	// Enable the generation of typescript schema during runtime, this is disabled by default to prevent unnecessary overhead when not needed
	EnableSchemaGeneration bool

	// Enable debug mode to log useful info
	EnableDebugMode bool

	// A function that will be called when an error occurs, it should ideally return a marshallable struct
	ErrorHandler ErrorHandler
}

// Robin is just going to be an adapter for something like Echo
func New(opts Options) *Robin {
	errorHandler := DefaultErrorHandler
	if opts.ErrorHandler != nil {
		errorHandler = opts.ErrorHandler
	}

	enableTSGen := opts.EnableSchemaGeneration
	// The environment variable takes precedence over whatver is set in code
	if v, isSet := os.LookupEnv(EnvRobinEnableTSGen); isSet {
		enableTSGen = strings.ToLower(v) == "true" || v == "1"
	}

	return &Robin{
		enableTypescriptGen: enableTSGen,
		cacheResolvedTypes:  opts.CacheResolvedTypes,
		debug:               opts.EnableDebugMode,
		procedures:          make(map[string]Procedure),
		errorHandler:        errorHandler,
	}
}

func (r *Robin) Add(procedure Procedure) *Robin {
	if _, ok := r.procedures[procedure.Name()]; ok {
		if r.debug {
			slog.Warn(
				"Attempted to add a duplicate procedure, skipping...",
				slog.String("procedureName", procedure.Name()),
			)
		}

		return r
	}

	r.procedures[procedure.Name()] = procedure
	return r
}

func (r *Robin) AddProcedure(procedure Procedure) *Robin {
	return r.Add(procedure)
}

func (r *Robin) Build() *Instance {
	return &Instance{bindingsPath: r.bindingsPath, robin: r}
}

func (r *Robin) serveHTTP(w http.ResponseWriter, req *http.Request) {
	defer func(r *Robin) {
		if e := recover(); e != nil {
			r.sendError(w, RobinError{Reason: fmt.Sprintf("Panic trapped: %v", e)})
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
		r.sendError(w, RobinError{Reason: "Procedure not found"})
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
		r.sendError(
			w,
			RobinError{
				Reason: "Invalid procedure type, expect one of 'query' or 'mutation', got " + string(
					ctx.ProcedureType,
				),
			},
		)
		return
	}
}

// TODO: split this into another function that handles errors from the robin handlers
func (r *Robin) sendError(w http.ResponseWriter, err error) {
	if r.debug {
		slog.Error("An error occurred in handler", slog.Any("error", err))
	}

	// TODO: the error handler should be able to return anything it wants and then we decide if we want to put it in a `data` field in the map struct or not
	errorResp, code := r.errorHandler(err)
	errMap := map[string]string{"error": string(errorResp)}
	jsonResp, err := json.Marshal(errMap)
	if err != nil {
		slog.Error("Failed to marshal error response", slog.String("error", err.Error()))
		w.WriteHeader(500)
		w.Write([]byte("Internal server error"))
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write([]byte(jsonResp))
}
