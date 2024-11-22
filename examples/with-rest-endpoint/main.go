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

var todos = []Todo{
	{"Hello world!", false, time.Now().Add(-time.Hour)},
	{"Hello world again!", true, time.Now()},
}

func main() {
	r, err := robin.New(robin.Options{
		EnableDebugMode: true,
		CodegenOptions: robin.CodegenOptions{
			Path:             ".",
			GenerateBindings: false,
		},
	})
	if err != nil {
		log.Fatalf("Failed to create a new Robin instance: %s", err)
	}

	i, err := r.
		Add(robin.Query("ping", ping)).
		Add(robin.Query("fail", fail)).
		Add(robin.Query("list-todos", listTodos)).
		Add(robin.Mutation("create.todo", createTodo)).
		Build()
	if err != nil {
		log.Fatalf("Failed to build Robin instance: %s", err)
	}

	if err := i.Serve(
		robin.ServeOptions{Port: 8060,
			Route: "/",
			RestApiOptions: &robin.RestApiOptions{
				Enable: true,
			},
		}); err != nil {
		log.Fatalf("Failed to serve Robin instance: %s", err)
		return
	}
}

func ping(ctx *robin.Context, _ robin.Void) (string, error) {
	return "pong", nil
}

func listTodos(ctx *robin.Context, _ robin.Void) ([]Todo, error) {
	return todos, nil
}

func createTodo(ctx *robin.Context, todo Todo) (Todo, error) {
	todo.CreatedAt = time.Now()
	todos = append(todos, todo)
	return todo, nil
}

// Yes, you can just return normal errors!
func fail(ctx *robin.Context, _ robin.Void) (robin.Void, error) {
	return robin.Void{}, errors.New("This is a procedure error!")
}
