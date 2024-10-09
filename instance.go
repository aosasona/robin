package robin

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"go.trulyao.dev/robin/generator"
)

type (
	Instance struct {
		// Knobs for typescript code generation
		codegenOptions *CodegenOptions

		// Internal pointer to the current robin instance
		robin *Robin

		// Port to run the server on
		port int

		// Route to run the robin handler on
		route string
	}

	CorsOptions struct {
		// Allowed origins
		Origins []string

		// Allowed headers
		Headers []string

		// Allowed methods
		Methods []string

		// Exposed headers
		ExposedHeaders []string

		// Allow credentials
		AllowCredentials bool

		// Max age
		MaxAge int

		// Preflight headers
		PreflightHeaders map[string]string
	}

	ServeOptions struct {
		// Port to run the server on
		Port int

		// Route to run the robin handler on
		Route string

		// CORS options
		CorsOptions *CorsOptions
	}
)

func preflight(w http.ResponseWriter, opts *CorsOptions) {
	if opts.PreflightHeaders != nil {
		for k, v := range opts.PreflightHeaders {
			w.Header().Set(k, v)
		}

		return
	}

	w.Header().Set("Access-Control-Allow-Origin", strings.Join(opts.Origins, ","))
	w.Header().
		Set("Access-Control-Allow-Headers", "Access-Control-Allow-Headers, Origin,Accept, X-Requested-With, Content-Type, Access-Control-Request-Method, Access-Control-Request-Headers")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
}

func cors(w http.ResponseWriter, opts *CorsOptions) {
	w.Header().Set("Access-Control-Allow-Origin", strings.Join(opts.Origins, ","))
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(opts.Headers, ","))
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(opts.Methods, ","))
	w.Header().Set("Access-Control-Expose-Headers", strings.Join(opts.ExposedHeaders, ","))
	w.Header().Set("Access-Control-Allow-Credentials", fmt.Sprintf("%t", opts.AllowCredentials))

	if opts.MaxAge > 0 {
		w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", opts.MaxAge))
	}
}

// Serve starts the robin server on the specified port
func (i *Instance) Serve(opts ...ServeOptions) error {
	corsOpts := &CorsOptions{
		Origins: []string{"*"},
		Headers: []string{"Content-Type", "Authorization"},
		Methods: []string{"POST", "OPTIONS"},
	}

	if len(opts) > 0 {
		optsPort := opts[0].Port
		if optsPort > 65535 {
			return errors.New("invalid port provided")
		}

		if optsPort > 0 {
			if optsPort < 1024 {
				slog.Warn("⚠️ Running robin on a privileged port", slog.Int("port", optsPort))
			}

			i.port = optsPort
		}

		i.route = strings.TrimSpace(strings.Trim(opts[0].Route, "/"))
		if opts[0].CorsOptions != nil {
			corsOpts = opts[0].CorsOptions
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /"+i.route, func(w http.ResponseWriter, r *http.Request) {
		cors(w, corsOpts)
		i.Handler()(w, r)
	})

	// Handle CORS preflight requests
	mux.HandleFunc("/"+i.route, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			preflight(w, corsOpts)
			return
		}
	})

	slog.Info(
		"📡 Robin server is listening",
		slog.Int("port", i.port),
		slog.String("route", "/"+i.route),
	)
	return http.ListenAndServe(fmt.Sprintf(":%d", i.port), mux)
}

// Handler returns the robin handler to be used with a custom (mux) router
func (i *Instance) Handler() http.HandlerFunc {
	return i.robin.serveHTTP
}

// SetPort sets the port to run the server on (default is 8081; to avoid conflicts with other services)
// WARNING: this only applies when calling `Serve()`, if you're using the default handler, you can set the port directly on the `http.Server` instance, you may have to update the client side to reflect the new port
func (i *Instance) SetPort(port int) {
	i.port = port
}

// SetRoute sets the route to run the robin handler on (default is `/_robin`)
// WARNING: this only applies when calling `Serve()`, if you're using the default handler, you can set the route using a mux router or similar, ensure that the client side reflects the new route
func (i *Instance) SetRoute(route string) {
	i.route = route
}

// Export exports the typescript schema (and bindings; if enabled) to the specified path
func (i *Instance) Export(optPath ...string) error {
	if !i.codegenOptions.GenerateSchema && !i.codegenOptions.GenerateBindings {
		return nil
	}

	// Figure out what path to use depending on user configurations
	path := i.codegenOptions.Path
	if len(optPath) > 0 {
		path = optPath[0]
	}

	// Ensure the path meets all out requirements
	if err := i.validatePath(path); err != nil {
		return err
	}

	// Generate the types
	g := generator.New(i.robin.procedures.List())
	schemaString, err := g.GenerateSchema()
	if err != nil {
		return err
	}

	// Write the schema to a file if it's enabled
	if i.codegenOptions.GenerateSchema && len(schemaString) > 0 {
		if err := i.writeSchemaToFile(path, strings.TrimSpace(schemaString)); err != nil {
			return err
		}
	}

	// Generate the methods if they're enabled and write them to a file
	if i.codegenOptions.GenerateBindings {
		bindingsString, err := g.GenerateBindings(generator.GenerateBindingsOpts{
			IncludeSchema:  !i.codegenOptions.GenerateSchema,
			Schema:         schemaString,
			UseUnionResult: i.codegenOptions.UseUnionResult,
			ThrowOnError:   i.codegenOptions.ThrowOnError,
		})
		if err != nil {
			return err
		}

		if err := i.writeBindingsToFile(path, bindingsString); err != nil {
			return err
		}
	}

	return nil
}

func (i *Instance) writeBindingsToFile(path, bindings string) error {
	filePath := fmt.Sprintf("%s/bindings.ts", path)

	if err := os.WriteFile(filePath, []byte(bindings), 0o644); err != nil {
		return fmt.Errorf("failed to write bindings to file: %s", err.Error())
	}

	slog.Info("📦 Typescript bindings exported successfully", slog.String("path", filePath))
	return nil
}

func (i *Instance) writeSchemaToFile(path, schema string) error {
	filePath := fmt.Sprintf("%s/schema.ts", path)

	if err := os.WriteFile(filePath, []byte(schema), 0o644); err != nil {
		return fmt.Errorf("failed to write schema to file: %s", err.Error())
	}

	slog.Info("📦 Typescript schema exported successfully", slog.String("path", filePath))
	return nil
}

func (i *Instance) validatePath(path string) error {
	if strings.TrimSpace(path) == "" {
		return errors.New(
			"no bindings export path provided, you can pass this to the `Export` method after calling `Build()` or as part of the opts during the instance creation with `robin.New(...)`",
		)
	}

	// Check that the path provided exists and is a directory
	stat, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat bindings path: %v", err)
	}

	if !stat.IsDir() {
		return errors.New("provided path is not a directory")
	}

	return nil
}
