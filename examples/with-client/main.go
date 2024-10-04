package main

import (
	"database/sql"
	"log/slog"

	"todo/handler"

	"go.trulyao.dev/robin"
	_ "go.trulyao.dev/seer"
)

// TODO: add SQlite
func main() {
	r, err := robin.New(robin.Options{
		CodegenOptions: robin.CodegenOptions{
			Path:             "./client",
			GenerateBindings: true,
		},
		EnableDebugMode: false,
	})
	if err != nil {
		slog.Error("Failed to create a new Robin instance", slog.String("error", err.Error()))
		return
	}

	// TODO: replace with actual database
	var db *sql.DB
	h := handler.New(db)

	i, err := r.
		Add(robin.Query("ping", h.Ping)).
		Add(robin.Query("todos.list", h.List)).
		Add(robin.Mutation("todos.create", h.Create)).
		Build()
	if err != nil {
		slog.Error("Failed to build Robin instance", slog.String("error", err.Error()))
		return
	}

	if err := i.Export(); err != nil {
		slog.Error("Failed to export client", slog.String("error", err.Error()))
		return
	}

	if err := i.Serve(); err != nil {
		slog.Error("Failed to serve Robin instance", slog.String("error", err.Error()))
		return
	}
}
