package robin

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"text/tabwriter"

	"go.trulyao.dev/robin/types"
)

type (
	RestEndpoint struct {
		// Name of the procedure
		ProcedureName string

		// Path to the endpoint e.g. /list.users
		Path string

		// HTTP method to use for the endpoint
		Method types.HttpMethod

		// HTTP method to use for the endpoint
		HandlerFunc http.HandlerFunc
	}

	Endpoints []*RestEndpoint
)

// String returns the string representation of the rest endpoint
func (re *RestEndpoint) String() string {
	str := strings.Builder{}

	w := tabwriter.NewWriter(&str, 0, 0, 1, ' ', 0)
	fmt.Fprintf(w, "%s\t[%s]\t%s", re.ProcedureName, re.Method, re.Path)

	_ = w.Flush()
	return str.String()
}

// String returns the string representation of the rest endpoints
func (e Endpoints) String() string {
	str := strings.Builder{}

	w := tabwriter.NewWriter(&str, 0, 0, 1, ' ', 0)
	fmt.Fprintf(w, "Procedure\tMethod\tPath\n")
	for _, endpoint := range e {
		fmt.Fprintf(
			w,
			"%s\t[%s]\t%s\n",
			endpoint.ProcedureName,
			endpoint.Method,
			endpoint.Path,
		)
	}

	_ = w.Flush()
	return str.String()
}

// BuildProcedureHttpHandler builds an http handler for the given procedure
func (i *Instance) BuildProcedureHttpHandler(procedure Procedure) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := types.NewContext(req, &w)

		ctx.SetProcedureName(procedure.Name())
		ctx.SetProcedureType(procedure.Type())

		if err := i.robin.handleProcedureCall(ctx, procedure); err != nil {
			i.robin.sendError(w, err)
			return
		}
	}
}

// BuildRestEndpoints builds the rest endpoints for the robin instance based on the procedures
//
// The prefix is used to prefix the path of the rest endpoints (e.g. /api/v1)
//
// This should be called after all the procedures have been added to the robin instance
func (i *Instance) BuildRestEndpoints(
	prefix string,
) Endpoints {
	var endpoints []*RestEndpoint

	prefix = trimUrlPath(prefix)

	for _, procedure := range i.robin.procedures.List() {
		method := types.HttpMethodGet
		if procedure.Type() == types.ProcedureTypeMutation {
			method = types.HttpMethodPost
		}

		alias := trimUrlPath(procedure.Alias())

		endpoint := &RestEndpoint{
			ProcedureName: procedure.Name(),
			Path:          trimUrlPath(fmt.Sprintf("/%s/%s", prefix, alias)),
			Method:        method,
			HandlerFunc:   i.BuildProcedureHttpHandler(procedure),
		}

		endpoints = append(endpoints, endpoint)
	}

	return endpoints
}

// AttachRestEndpoints attaches the RESTful endpoints to the provided mux router automatically
//
// NOTE: If you require more control, look at the `BuildRestEndpoints` and the `BuildProcedureHttpHandler` methods on the `Robin` instance
func (i *Instance) AttachRestEndpoints(mux *http.ServeMux, opts *RestApiOptions) {
	if opts == nil || !opts.Enable {
		slog.Warn("🔗 RESTful endpoints are disabled or not configured")
		return
	}

	prefix := strings.Trim(opts.Prefix, "/")
	if prefix == "" {
		prefix = "/api"
	}

	endpoints := i.BuildRestEndpoints(prefix)
	for _, endpoint := range endpoints {
		if i.robin.Debug() {
			slog.Info("🔗 Attaching RESTful endpoint", slog.String("endpoint", endpoint.String()))
		}

		mux.HandleFunc(fmt.Sprintf("%s %s", endpoint.Method, endpoint.Path), endpoint.HandlerFunc)
	}

	// Attach the not found handler
	if !opts.DisableNotFoundHandler {
		mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
			i.robin.sendError(w, types.NewError("Resource not found", http.StatusNotFound))
		})
	}

	// If debug is enabled, print the rest endpoints
	if i.robin.Debug() {
		fmt.Println("+------------------------------------+")
		fmt.Println("🔗 RESTful endpoints")
		fmt.Println("+------------------------------------+")
		fmt.Println(endpoints.String())
	}
}
