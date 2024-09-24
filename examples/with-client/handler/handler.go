package handler

import (
	"database/sql"

	"go.trulyao.dev/robin"
)

// TODO: take in sqlc query interface here instead
type handler struct {
	db *sql.DB
}

func New(db *sql.DB) *handler {
	return &handler{db}
}

// TODO: fix the stubs
func (h *handler) List(ctx *robin.Context, _ robin.Void) ([]string, error) {
	return []string{}, nil
}

func (h *handler) Create(ctx *robin.Context, _ robin.Void) (string, error) {
	return "", nil
}
