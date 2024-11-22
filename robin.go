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
	Middleware    = types.Middleware
)

// Re-exported constants
const (
	ProcedureTypeQuery    ProcedureType = types.ProcedureTypeQuery
	ProcedureTypeMutation ProcedureType = types.ProcedureTypeMutation
)

const (
	ProcSeparator = "__"
	ProcNameKey   = ProcSeparator + "proc"

	// Environment variables to control code generation outside of the code
	EnvEnableSchemaGen   = "ROBIN_ENABLE_SCHEMA_GEN"
	EnvEnableBindingsGen = "ROBIN_ENABLE_BINDINGS_GEN"
)

var (
	// Valid procedure name regex
	ReValidProcedureName = regexp.MustCompile(`(?m)^([a-zA-Z0-9]+)([_\.\-]?[a-zA-Z0-9]+)+$`)

	// Invalid characters in a procedure name
	ReAlphaNumeric = regexp.MustCompile(`[^a-zA-Z0-9]+`)

	// Multiple dots in a procedure name
	ReIllegalDot = regexp.MustCompile(`\.{2,}`)

	// Valid/common words associated with queries
	ReQueryWords = regexp.MustCompile(`(?i)(^(get|fetch|list|lookup|search|find|query|retrieve|show|view|read)\.)`)

	ReMutationWords = regexp.MustCompile(`(?i)(^(create|add|insert|update|upsert|edit|modify|change|delete|remove|destroy)\.)`)
)

type (
	CodegenOptions struct {
		// Path to the generated folder for typescript bindings and/or schema
		Path string

		// Whether to generate the typescript bindings or not.
		//
		// NOTE: You can simply generate the schema without the bindings by enabling `GenerateSchema` and disabling this
		//
		// WARNING: If this is enabled and `GenerateSchema` is disabled, the schema will be generated as part of the bindings in the same file
		GenerateBindings bool

		// Whether to generate the typescript schema separately or not
		GenerateSchema bool

		// Whether to use the union result type or not - when enabled, the result type will be a uniion of the Ok and Error types which would disallow access to any of the fields without checking the `ok` field first
		UseUnionResult bool

		// Whether to throw a ProcedureCallError when a procedure call fails for any reason (e.g. invalid payload, user-defined error, etc.) instead of returning an error result
		ThrowOnError bool
	}

	Options struct {
		// Options for controlling code generation
		CodegenOptions CodegenOptions

		// Enable debug mode to log useful info
		EnableDebugMode bool

		// A function that will be called when an error occurs, it should ideally return a marshallable struct
		ErrorHandler ErrorHandler
	}

	GlobalMiddleware struct {
		Name string
		Fn   Middleware
	}

	Robin struct {
		// Controls Typescript code generation
		codegenOptions CodegenOptions

		// Enable debug mode to log useful info
		debug bool

		// A list of query and mutation procedures
		procedures *Procedures

		// A map of global middleware that will be executed before any procedure is called unless explicitly excluded/opted out of
		// NOTE: a slice has been used instead of a map to maintain the order of insertion as this is crucial to the order of execution for some middlewares
		namedGlobalMiddleware []GlobalMiddleware

		// A function that will be called when an error occurs, if not provided, the default error handler will be used
		errorHandler ErrorHandler
	}
)

// Robin is just going to be an adapter for something like Echo
func New(opts Options) (*Robin, error) {
	robin := new(Robin)

	errorHandler := DefaultErrorHandler
	if opts.ErrorHandler != nil {
		errorHandler = opts.ErrorHandler
	}

	codegenOptions, err := robin.extractCodegenOptions(&opts)
	if err != nil {
		return nil, err
	}

	robin = &Robin{
		codegenOptions: codegenOptions,
		debug:          opts.EnableDebugMode,
		procedures:     &Procedures{},
		errorHandler:   errorHandler,
	}

	return robin, nil
}

// Add a new procedure to the Robin instance
// If a procedure with the same name already exists, it will be skipped and a warning will be logged in debug mode
func (r *Robin) Add(procedure Procedure) *Robin {
	if r.debug {
		slog.Info("Adding procedure", slog.String("procedureName", procedure.Name()))
	}

	if r.procedures.Exists(procedure.Name(), procedure.Type()) {
		if r.debug {
			slog.Warn(
				"Attempted to add a duplicate procedure, skipping...",
				slog.String("procedureName", procedure.Name()),
			)
		}

		return r
	}

	r.procedures.Add(procedure)
	return r
}

// Add a new procedure to the Robin instance - an alias for `Add`
func (r *Robin) AddProcedure(procedure Procedure) *Robin {
	return r.Add(procedure)
}

// Use adds a global middleware to the robin instance, these middlewares will be executed before any procedure is called unless explicitly excluded/opted out of
// The order in which the middlewares are added is the order in which they will be executed before the procedures
//
// WARNING: Global middlewares are ALWAYS executed before the procedure's middleware functions
//
// NOTE: Use `procedure.ExcludeMiddleware(...)` to exclude a middleware from a specific procedure
func (r *Robin) Use(name string, middleware Middleware) *Robin {
	r.namedGlobalMiddleware = append(r.namedGlobalMiddleware, GlobalMiddleware{Name: name, Fn: middleware})
	return r
}

// Build the Robin instance
func (r *Robin) Build() (*Instance, error) {
	// Validate all procedures
	for _, procedure := range *r.procedures {
		if err := procedure.Validate(); err != nil {
			return nil, err
		}

		if r.debug {
			slog.Info("Procedure validated", slog.String("procedureName", procedure.Name()))
		}

		// Check if we have excluded a wildcard middleware
		if procedure.ExcludedMiddleware().Has("*") {
			continue
		}

		var globalMiddleware []Middleware // This is to maintain the order of execution, attempting to prepending in the loop will reverse the order
		// Add global middleware to the procedures
		for _, middleware := range r.namedGlobalMiddleware {
			if procedure.ExcludedMiddleware().Has(middleware.Name) {
				continue
			}

			globalMiddleware = append(globalMiddleware, middleware.Fn)
		}

		// Prepend global middleware to the procedure's middleware chain
		procedure.PrependMiddleware(globalMiddleware...)

		if r.debug {
			slog.Info("Global middleware added to procedure", slog.String("procedureName", procedure.Name()), slog.Int("middlewareCount", len(globalMiddleware)))
		}

		procedure.ExcludedMiddleware().Clear() // Clear the exclusion list to free up memory taken from the dedup
	}

	if r.debug {
		slog.Info("Robin instance built successfully", slog.String("procedures", fmt.Sprintf("%v", r.procedures.Keys())))
	}

	return &Instance{
		codegenOptions: &r.codegenOptions,
		robin:          r,
		port:           8081,
		route:          "_robin",
	}, nil
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

	ctx := types.NewContext(req, &w)
	procedureType, procedureName, err := r.getProcedureMetaFromURL(req.URL)
	if err != nil {
		r.sendError(w, err)
		return
	}

	ctx.SetProcedureName(procedureName)
	ctx.SetProcedureType(procedureType)

	procedure, found := r.findProcedure(ctx.ProcedureName(), ctx.ProcedureType())
	if !found {
		r.sendError(
			w,
			r.makeMissingProcedureError(ctx.ProcedureName(), ctx.ProcedureType()),
		)
		return
	}

	switch ProcedureType(ctx.ProcedureType()) {
	case ProcedureTypeQuery, ProcedureTypeMutation:
		err := r.handleProcedureCall(ctx, procedure)
		if err != nil {
			r.sendError(w, err)
			return
		}

	default:
		r.sendError(
			w,
			types.RobinError{
				Reason: fmt.Sprintf(
					"Invalid procedure type, expect one of 'query' or 'mutation', got %s",
					string(ctx.ProcedureType()),
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

	for _, procedure := range *r.procedures {
		distance := levenshtein.ComputeDistance(
			strings.ToLower(procedure.Name()),
			strings.ToLower(procedureName),
		)

		if closest == nil || distance < closestDistance {
			closest = procedure
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

// Utility function to send an error response
func (r *Robin) sendError(w http.ResponseWriter, err error) {
	if r.debug {
		slog.Error("An error occurred in handler", slog.Any("error", err))
	}

	errorResponse, code := r.errorHandler(err)
	errMap := map[string]any{"error": errorResponse, "ok": false}
	jsonResp, err := json.Marshal(errMap)
	if err != nil {
		slog.Error("Failed to marshal error response", slog.String("error", err.Error()))
		w.WriteHeader(500)
		_, _ = w.Write([]byte("Internal server error"))
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write([]byte(jsonResp))
}

// extractCodegenOptions extracts the codegen options from the provided options and environment variables
func (r *Robin) extractCodegenOptions(opts *Options) (CodegenOptions, error) {
	enableSchemaGen := opts.CodegenOptions.GenerateSchema
	// The environment variable takes precedence over whatver is set in code
	if v, ok := os.LookupEnv(EnvEnableSchemaGen); ok {
		enableSchemaGen = strings.ToLower(v) == "true" || v == "1"
	}

	enableBindingsGen := opts.CodegenOptions.GenerateBindings
	if v, ok := os.LookupEnv(EnvEnableBindingsGen); ok {
		enableBindingsGen = strings.ToLower(v) == "true" || v == "1"
	}

	// Ensure the bindings path is a valid directory
	if opts.CodegenOptions.Path != "" &&
		(opts.CodegenOptions.GenerateBindings || opts.CodegenOptions.GenerateSchema) {
		if _, err := os.Stat(opts.CodegenOptions.Path); os.IsNotExist(err) {
			slog.Warn(
				"Provided bindings path does not exist, creating it...",
				slog.String("path", opts.CodegenOptions.Path),
			)

			err := os.MkdirAll(opts.CodegenOptions.Path, 0o755)
			if err != nil {
				return CodegenOptions{}, fmt.Errorf("failed to create bindings path: %v", err)
			}
		}
	}

	return CodegenOptions{
		Path:             opts.CodegenOptions.Path,
		GenerateBindings: enableBindingsGen,
		GenerateSchema:   enableSchemaGen,
		UseUnionResult:   opts.CodegenOptions.UseUnionResult,
		ThrowOnError:     opts.CodegenOptions.ThrowOnError,
	}, nil
}
