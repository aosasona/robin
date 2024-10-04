package handler

import (
	"database/sql"
	"fmt"

	"go.trulyao.dev/robin"
)

// TODO: use bbolt.DB
type handler struct {
	db *sql.DB
}

func New(db *sql.DB) *handler {
	return &handler{db}
}

func (h *handler) Ping(ctx *robin.Context, data string) (string, error) {
	return fmt.Sprintf("Pong: %s", data), nil
}

// TODO: fix the stubs
func (h *handler) List(ctx *robin.Context, _ robin.Void) ([]string, error) {
	return []string{}, nil
}

func (h *handler) Create(ctx *robin.Context, _ robin.Void) (string, error) {
	return "", nil
}
