package handler

import (
	"go.etcd.io/bbolt"
	"go.trulyao.dev/robin"
)

type handler struct {
	db *bbolt.DB
}

type CreateInput struct {
	Title string `json:"title"`
}

func New(db *bbolt.DB) *handler {
	return &handler{db}
}

// TODO: update stubs
func (h *handler) Ping(ctx *robin.Context, data string) (string, error) {
	return "Hey", nil
}

func (h *handler) List(ctx *robin.Context, _ robin.Void) ([]string, error) {
	return []string{"Hello, world!"}, nil
}

func (h *handler) Create(ctx *robin.Context, input CreateInput) (CreateInput, error) {
	return input, nil
}
