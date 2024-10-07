package main

import (
	"errors"
	"log"
	"time"

	"go.trulyao.dev/robin"
)

type Todo struct {
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

func main() {
	r, err := robin.New(robin.Options{
		CodegenOptions: robin.CodegenOptions{
			Path:             ".",
			GenerateBindings: true,
			ThrowOnError:     true,
			UseUnionResult:   true,
		},
	})
	if err != nil {
		log.Fatalf("Failed to create a new Robin instance: %s", err)
	}

	i, err := r.
		Add(robin.Query("ping", ping)).
		Add(robin.Query("fail", fail)).
		Add(robin.Query("todos.list", listTodos)).
		Add(robin.Mutation("todos.create", createTodo)).
		Build()
	if err != nil {
		log.Fatalf("Failed to build Robin instance: %s", err)
	}

	if err := i.Export(); err != nil {
		log.Fatalf("Failed to export client: %s", err)
	}

	if err := i.Serve(robin.ServeOptions{Port: 8060, Route: "/"}); err != nil {
		log.Fatalf("Failed to serve Robin instance: %s", err)
		return
	}
}

func ping(ctx *robin.Context, _ robin.Void) (string, error) {
	return "pong", nil
}

func listTodos(ctx *robin.Context, _ robin.Void) ([]Todo, error) {
	return []Todo{
		{"Hello world!", false, time.Now()},
		{"Hello world again!", true, time.Now()},
	}, nil
}

func createTodo(ctx *robin.Context, todo Todo) (Todo, error) {
	todo.CreatedAt = time.Now()
	return todo, nil
}

// Yes, you can just return normal errors!
func fail(ctx *robin.Context, _ robin.Void) (robin.Void, error) {
	return robin.Void{}, errors.New("This is a procedure error!")
}
