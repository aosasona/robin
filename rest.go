package robin

import (
	"fmt"
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
func (r *Robin) BuildProcedureHttpHandler(
	procedure Procedure,
) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := types.NewContext(req, &w)

		ctx.SetProcedureName(procedure.Name())
		ctx.SetProcedureType(procedure.Type())

		if err := r.handleProcedureCall(ctx, procedure); err != nil {
			r.sendError(w, err)
			return
		}
	}
}

// BuildRestEndpoints builds the rest endpoints for the robin instance based on the procedures
//
// The prefix is used to prefix the path of the rest endpoints (e.g. /api/v1)
//
// This should be called after all the procedures have been added to the robin instance
func (r *Robin) BuildRestEndpoints(
	prefix string,
) Endpoints {
	var endpoints []*RestEndpoint

	prefix = strings.Trim(prefix, "/")

	for _, procedure := range *r.procedures {
		method := types.HttpMethodGet
		if procedure.Type() == types.ProcedureTypeMutation {
			method = types.HttpMethodPost
		}

		endpoint := &RestEndpoint{
			ProcedureName: procedure.Name(),
			Path:          fmt.Sprintf("/%s/%s", prefix, procedure.Alias()),
			Method:        method,
			HandlerFunc:   r.BuildProcedureHttpHandler(procedure),
		}

		endpoints = append(endpoints, endpoint)
	}

	return endpoints
}
