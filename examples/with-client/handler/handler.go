package handler

import (
	"fmt"

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

func (h *handler) Ping(ctx *robin.Context, data string) (string, error) {
	return "Hey", nil
}

// TODO: fix the stubs
func (h *handler) List(ctx *robin.Context, _ robin.Void) ([]string, error) {
	return []string{"Hello, world!"}, nil
}

func (h *handler) Create(ctx *robin.Context, input CreateInput) (CreateInput, error) {
	fmt.Println("Creating todo with title:", input.Title)
	return input, nil
}
