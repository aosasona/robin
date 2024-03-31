package robin

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
)

func (r *Robin) handleProcedureCall(ctx *Context, procedure Procedure) error {
	var (
		data = struct {
			Payload any `json:"d"`
		}{Payload: procedure.PayloadInterface()}

		response map[string]interface{} = make(map[string]interface{})
	)

	response["data"] = nil

	if err := json.NewDecoder(ctx.Request.Body).Decode(&data); err != nil {
		if r.debug && err.Error() != "EOF" {
			slog.Error("Failed to decode request body", slog.String("error", err.Error()))
		}

		// CAVEAT: sending an empty body can cause a panic here
		if err = invalidTypesError(procedure.PayloadInterface(), data.Payload); err != nil {
			return err
		}
	}

	result, err := procedure.Call(ctx, data.Payload)
	if err != nil {
		return err
	}

	if result != nil {
		response["data"] = result
	}

	strResponse, err := json.Marshal(response)
	if err != nil {
		return InternalError{Reason: "Failed to marshal response", OriginalError: err}
	}

	ctx.Response.Header().Add("content-type", "application/json")
	ctx.Response.WriteHeader(200)
	ctx.Response.Write(strResponse)

	return nil
}

func (r *Robin) getProcedureMetaFromURL(url *url.URL) (ProcedureType, string, error) {
	var (
		procedureName string
		procedureType ProcedureType
	)

	// Queries can only be issued via GET requests, and Mutations can only be issued via POST requests but in both cases, the procedure name is attached to the URL query
	proc := url.Query().Get(ProcNameKey)
	if strings.TrimSpace(proc) == "" {
		return "", "", errors.New("No procedure name provided")
	}

	procParts := strings.Split(proc, ProcSeparator)
	if len(procParts) != 2 {
		return "", "", fmt.Errorf("Invalid procedure param provided in URL, expected format (q|m)%s[name] e.g q%sgetUser", ProcSeparator, ProcSeparator)
	}

	shortProcType := procParts[0]
	switch shortProcType {
	case "q":
		procedureType = ProcedureTypeQuery
	case "m":
		procedureType = ProcedureTypeMutation
	default:
		return "", "", errors.New("No procedure name provided")
	}

	procedureName = procParts[1]

	return procedureType, procedureName, nil
}

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