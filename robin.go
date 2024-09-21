package robin

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/agnivade/levenshtein"
	"go.trulyao.dev/robin/types"
)

// Re-exported types
type (
	Void = types.Void

	Error      = types.Error
	RobinError = types.RobinError

	ProcedureType = types.ProcedureType
	Procedure     = types.Procedure
	Context       = types.Context
)

// Re-exported constants
const (
	ProcedureTypeQuery    ProcedureType = types.ProcedureTypeQuery
	ProcedureTypeMutation ProcedureType = types.ProcedureTypeMutation
)

const (
	ProcSeparator = "__"
	ProcNameKey   = ProcSeparator + "proc"

	EnvRobinEnableTSGen = "ROBIN_ENABLE_SCHEMA_GEN"
)

var procedureNameRegex = regexp.MustCompile(`(?m)[^a-zA-Z0-9]`)

type (
	Robin struct {
		// Path to the generated typescript schema
		bindingsPath string

		// Enable the generation of typescript schema during runtime, this is disabled by default to prevent unnecessary overhead when not needed
		enableTypescriptGen bool

		// Enable debug mode to log useful info
		debug bool

		// A list of query and mutation procedures
		procedures map[string]Procedure

		// A function that will be called when an error occurs, if not provided, the default error handler will be used
		errorHandler ErrorHandler
	}
)

// TODO: add codegen struct to contain flags for generating just the schema or the schema and the typescript bindings
type Options struct {
	// Path to the generated folder for typescript bindings
	BindingsPath string

	// Enable the generation of typescript schema during runtime, this is disabled by default to prevent unnecessary overhead when not needed
	EnableSchemaGeneration bool

	// Enable debug mode to log useful info
	EnableDebugMode bool

	// A function that will be called when an error occurs, it should ideally return a marshallable struct
	ErrorHandler ErrorHandler
}

// Robin is just going to be an adapter for something like Echo
func New(opts Options) (*Robin, error) {
	robin := new(Robin)

	errorHandler := DefaultErrorHandler
	if opts.ErrorHandler != nil {
		errorHandler = opts.ErrorHandler
	}

	enableTSGen := opts.EnableSchemaGeneration
	// The environment variable takes precedence over whatver is set in code
	if v, isSet := os.LookupEnv(EnvRobinEnableTSGen); isSet {
		enableTSGen = strings.ToLower(v) == "true" || v == "1"
	}

	// Ensure the bindings path is a valid directory
	if opts.BindingsPath != "" {
		if _, err := os.Stat(opts.BindingsPath); os.IsNotExist(err) {
			slog.Warn(
				"Provided bindings path does not exist, creating it...",
				slog.String("path", opts.BindingsPath),
			)

			err := os.MkdirAll(opts.BindingsPath, 0o755)
			if err != nil {
				return nil, fmt.Errorf("failed to create bindings path: %v", err)
			}
		}
	}

	robin = &Robin{
		bindingsPath:        opts.BindingsPath,
		enableTypescriptGen: enableTSGen,
		debug:               opts.EnableDebugMode,
		procedures:          make(map[string]Procedure),
		errorHandler:        errorHandler,
	}

	return robin, nil
}

// Add a new procedure to the Robin instance
// If a procedure with the same name already exists, it will be skipped and a warning will be logged in debug mode
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

// Add a new procedure to the Robin instance - an alias for `Add`
func (r *Robin) AddProcedure(procedure Procedure) *Robin {
	return r.Add(procedure)
}

// Build the Robin instance
func (r *Robin) Build() *Instance {
	return &Instance{
		enableTypescriptGen: r.enableTypescriptGen,
		robin:               r,
		bindingsPath:        r.bindingsPath,
		port:                8081,
		route:               "_robin",
	}
}

// serveHTTP is the main handler for all incoming HTTP requests
// It takes the request, and transforms it into a Robin Context, then calls the appropriate procedure if present
func (r *Robin) serveHTTP(w http.ResponseWriter, req *http.Request) {
	defer func(r *Robin) {
		if e := recover(); e != nil {
			r.sendError(w, types.RobinError{Reason: fmt.Sprintf("Panic trapped: %v", e)})
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
		r.sendError(
			w,
			r.makeMissingProcedureError(ctx.ProcedureName, ctx.ProcedureType),
		)
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
			types.RobinError{
				Reason: "Invalid procedure type, expect one of 'query' or 'mutation', got " + string(
					ctx.ProcedureType,
				),
			},
		)
		return
	}
}

// makeMissingProcedureError creates an error that indicates that a procedure was not found and suggests the closest procedure name if any
func (r *Robin) makeMissingProcedureError(procedureName string, procedureType ProcedureType) error {
	var (
		closest         Procedure
		closestDistance int
		errString       = fmt.Sprintf(
			"Procedure `%s` (%s) not found",
			procedureName,
			procedureType,
		)
	)

	for name, proc := range r.procedures {
		distance := levenshtein.ComputeDistance(
			strings.ToLower(name),
			strings.ToLower(procedureName),
		)

		if closest == nil || distance < closestDistance {
			closest = proc
			closestDistance = distance
		}
	}

	if closest != nil {
		errString = fmt.Sprintf(
			"Procedure `%s` (%s) not found, did you mean `%s` (%s)?",
			procedureName,
			procedureType,
			closest.Name(),
			closest.Type(),
		)
	}

	return types.RobinError{Reason: errString}
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
