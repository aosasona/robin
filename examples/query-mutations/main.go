package main

import (
	"go.trulyao.dev/robin"
)

type User struct {
	Name string
}

type Error struct {
	Message string
}

var users = []User{
	{Name: "John Doe"},
}

func main() {
	errorHandler := func(err error) any {
		return Error{Message: err.Error()}
	}

	robin.
		New(errorHandler).
		Add(robin.Query("ping", func(ctx robin.Context, _ string) (string, error) {
			return "pong", nil
		})).
		Add(robin.Query("users", func(ctx robin.Context, _ string) ([]User, error) {
			return users, nil
		}))
}
