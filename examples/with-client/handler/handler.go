package handler

import (
	"go.etcd.io/bbolt"
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
