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

	i := r.Add(robin.Query("list", h.List)).
		Add(robin.Mutation("create", h.Create)).
		Build()

	if err := i.ExportClient(); err != nil {
		slog.Error("Failed to export client", slog.String("error", err.Error()))
		return
	}

	i.Serve()
}
