package handler

import "go.trulyao.dev/robin"

func (h *handler) List(ctx *robin.Context, _ robin.Void) ([]string, error) {
	return []string{"Hello, world!"}, nil
}

func (h *handler) Create(ctx *robin.Context, input CreateInput) (CreateInput, error) {
	return input, nil
}
