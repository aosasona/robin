package robin

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"go.trulyao.dev/robin/internal/guarded"
)

func (r *Robin) handleProcedureCall(ctx *Context, procedure Procedure) error {
	var (
		data = struct {
			Payload any `json:"d"`
		}{Payload: procedure.PayloadInterface()}

		response = make(map[string]interface{})
	)

	response["data"] = nil

	if err := json.NewDecoder(ctx.Request.Body).Decode(&data); err != nil {
		if r.debug && err.Error() != "EOF" {
			slog.Error("Failed to decode request body", slog.String("error", err.Error()))
		}

		// CAVEAT: sending an empty body can cause a panic here
		if err = guarded.MakeCastError(procedure.PayloadInterface(), data.Payload); err != nil {
			return err
		}
	}

	// Call the procedure
	result, err := procedure.Call(ctx, data.Payload)
	if err != nil {
		return err
	}

	// TODO: Allow returning nil in the case of procedures that don't return anything
	if result != nil {
		response["data"] = result
	}

	strResponse, err := json.Marshal(response)
	if err != nil {
		return RobinError{Reason: "Failed to marshal response", OriginalError: err}
	}

	ctx.Response.Header().Add("content-type", "application/json")
	ctx.Response.WriteHeader(200)
	ctx.Response.Write(strResponse)

	return nil
}

// TODO: document
func (r *Robin) getProcedureMetaFromURL(url *url.URL) (ProcedureType, string, error) {
	var (
		procedureName string
		procedureType ProcedureType
	)

	// Queries can only be issued via GET requests, and Mutations can only be issued via POST requests but in both cases, the procedure name is attached to the URL query
	proc := url.Query().Get(ProcNameKey)
	if strings.TrimSpace(proc) == "" {
		return "", "", errors.New("no procedure name provided")
	}

	procParts := strings.Split(proc, ProcSeparator)
	if len(procParts) != 2 {
		return "", "", fmt.Errorf(
			"invalid procedure param provided in URL, expected format (q|m)%s[name] e.g q%sgetUser",
			ProcSeparator,
			ProcSeparator,
		)
	}

	shortProcType, procedureName := procParts[0], procParts[1]

	switch shortProcType {
	case "q":
		procedureType = ProcedureTypeQuery
	case "m":
		procedureType = ProcedureTypeMutation
	default:
		return "", "", errors.New("no procedure name provided")
	}

	return procedureType, procedureName, nil
}

// TODO: document
func (r *Robin) findProcedure(name string, procedureType ProcedureType) (*Procedure, bool) {
	procedure, ok := r.procedures[name]
	if !ok {
		return nil, false
	}

	if procedure.Type() != procedureType {
		return nil, false
	}

	return &procedure, true
}
