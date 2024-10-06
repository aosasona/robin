package main

import (
	"log"
	"log/slog"

	"todo/handler"
	"todo/utils"

	"go.etcd.io/bbolt"
	"go.trulyao.dev/robin"
	_ "go.trulyao.dev/seer"
)

func errorHandler(err error) (robin.Serializable, int) {
	message := err.Error()
	code := 500

	if e, ok := err.(utils.Error); ok {
		code = e.Code
		message = e.Message
	} else if e, ok := err.(robin.Error); ok {
		code = e.Code
		message = "An error occurred"
		slog.Error("An internal error occured", slog.String("message", message))
	}

	return robin.ErrorString(message), code
}

func main() {
	r, err := robin.New(robin.Options{
		CodegenOptions: robin.CodegenOptions{
			Path:             "./client",
			GenerateBindings: true,
			GenerateSchema:   false,
			UseUnionResult:   false,
		},
		EnableDebugMode: false,
		ErrorHandler:    errorHandler,
	})
	if err != nil {
		log.Fatalf("Failed to create a new Robin instance: %s", err)
	}

	db, err := bbolt.Open("todos.db", 0o666, nil)
	if err != nil {
		log.Fatalf("Failed to open BoltDB: %s", err)
		return
	}

	h := handler.New(db)

	i, err := r.
		Add(robin.Query("ping", h.Ping)).
		Add(robin.Query("todos.list", h.List)).
		Add(robin.Mutation("todos.create", h.Create)).
		Build()
	if err != nil {
		log.Fatalf("Failed to build Robin instance: %s", err)
	}

	if err := i.Export(); err != nil {
		log.Fatalf("Failed to export client: %s", err)
	}

	if err := i.Serve(); err != nil {
		log.Fatalf("Failed to serve Robin instance: %s", err)
		return
	}
}
