package robin

import (
	"net/http"
)

type Context struct {
	// The raw incoming request
	Request *http.Request

	// The raw response writer
	Response http.ResponseWriter

	// The name of the procedure
	ProcedureName string

	// The type of the procedure
	ProcedureType ProcedureType
}

// TODO: add common methods for the Context here
