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

type Instance struct {
	// Knobs for typescript code generation
	codegenOptions *CodegenOptions

	// Internal pointer to the current robin instance
	robin *Robin

	// Path to the generated folder for typescript bindings
	bindingsPath string

	// Port to run the server on
	port int

	// Route to run the robin handler on
	route string
}

type ServeOptions struct {
	// Port to run the server on
	Port int

	// Route to run the robin handler on
	Route string
}

// Serve starts the robin server on the specified port
func (i *Instance) Serve(opts ...ServeOptions) error {
	if len(opts) > 0 {
		i.port = opts[0].Port
		i.route = strings.TrimSpace(strings.Trim(opts[0].Route, "/"))
	}

	i.port = i.port
	i.route = i.route

	mux := http.NewServeMux()
	mux.Handle("POST /"+i.route, i.Handler())

	slog.Info("📡 Robin server is listening", slog.Int("port", i.port))
	http.ListenAndServe(fmt.Sprintf(":%d", i.port), mux)

	return nil
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

// ExportTSBindings exports the typescript bindings to the specified path
func (i *Instance) ExportTSBindings(optPath ...string) error {
	if !i.codegenOptions.GenerateSchema && !i.codegenOptions.GenerateBindings {
		return nil
	}

	// Automatically enable schema generation if bindings are enabled
	if !i.codegenOptions.GenerateSchema && i.codegenOptions.GenerateBindings {
		i.codegenOptions.GenerateSchema = true
	}

	// Figure out what path to use depending on user configurations
	path := i.bindingsPath
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

	// Write the schema to a file
	if err := i.writeSchemaToFile(path, strings.TrimSpace(schemaString)); err != nil {
		return err
	}

	// TODO: check if bindings generation is enabled and generate the bindings

	return nil
}

func (i *Instance) writeSchemaToFile(path, schema string) error {
	filePath := fmt.Sprintf("%s/schema.ts", path)

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create schema file: %s", err.Error())
	}
	defer file.Close()

	if _, err := file.WriteString(schema); err != nil {
		return fmt.Errorf("failed to write schema to file: %s", err.Error())
	}

	slog.Info("📦 Typescript bindings exported successfully", slog.String("path", filePath))
	return nil
}

func (i *Instance) validatePath(path string) error {
	if strings.TrimSpace(path) == "" {
		return errors.New(
			"no bindings export path provided, you can pass this to the `ExportTSBindings` method after calling `Build()` or as part of the opts during the instance creation with `robin.New(...)`",
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
