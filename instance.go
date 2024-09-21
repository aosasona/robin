package robin

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"go.trulyao.dev/robin/generator"
	"go.trulyao.dev/robin/types"
)

type Instance struct {
	// Enable generation of typescript bindings
	enableTypescriptGen bool

	// Path to the generated folder for typescript bindings
	bindingsPath string

	// Internal pointer to the current robin instance
	robin *Robin

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
	if !i.enableTypescriptGen {
		return nil
	}
	path := i.bindingsPath

	if len(optPath) > 0 {
		path = optPath[0]
	}

	if strings.TrimSpace(path) == "" {
		return errors.New(
			"no bindings export path provided, you can pass this to the `ExportTSBindings` method after calling `Build()` or as part of the opts during the instance creation with `robin.New(...)`",
		)
	}

	// TODO: check that is is a directory

	procedures := make([]types.Procedure, 0, len(i.robin.procedures))
	for _, p := range i.robin.procedures {
		procedures = append(procedures, p)
	}

	g := generator.New(procedures)
	types, err := g.GenerateSchema()
	if err != nil {
		return err
	}

	fmt.Println("==> TYPES\n" + types)

	return nil
}
