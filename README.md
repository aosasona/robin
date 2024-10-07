# Robin

> [!WARNING]
> The repo you are currently looking at contains an alpha-ish release and is not the right tool for you if you are not willing to put up with breaking changes from time to time and barebones documentation.
> Eventually, I intend to take the learnings from this duct-taped version to figure out the appropriate APIs and then rewrite to focus on a cleaner (and honestly, saner) code. I do not have the time to work on this heavily right now, **use at your own risk**.

# Introduction

Robin is an experimental and new(-ish) way to rapidly develop web applications in Go, based on another project; [mirror](https://github.com/aosasona/mirror).

It aims to provide an experience similar to those available in other langauges like [Rust (rspc)](https://rspc.dev) and [TypeScript (trpc)](https://trpc.io); allowing you to move fast without worrying about writing code to handle HTTP calls, data marshalling and unmarshalling, type definitions etc. while keeping both the server and client contracts in sync. Enough said, let's see some code.

# Installation

You can add robin directly in your project using the command below:

```sh
go get -u go.trulyao.dev/robin
```

# Example

## Server (Go)

Defining your procedures in the Go application/server is as simple as creating functions as you normally would (with a few known and unknown limitations) as shown below.

```go
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
```

## Client (TypeScript)

This is how you would use the generated client code in your TypeScript project.

```typescript
import Client from "./bindings.ts";

const client = Client.new({
	endpoint: "http://localhost:8060",
});

await client.queries.ping();

const todos = await client.queries.todosList();
const newTodo = await client.mutations.todosCreate({
	title: "Buy milk",
	completed: false,
});

console.log("todos -> ", todos);
console.log("newTodo -> ", newTodo);

// This should throw since the generated client is set to throw on errors
await client.queries.fail();
```

Running the usage script will yield this:

```sh
bun ./usage.ts
```

```text
todos ->  [
  {
    title: "Hello world!",
    completed: false,
    created_at: "2024-10-07T15:30:10.785946+01:00",
  }, {
    title: "Hello world again!",
    completed: true,
    created_at: "2024-10-07T15:30:10.785946+01:00",
  }
]
newTodo ->  {
  title: "Buy milk",
  completed: false,
  created_at: "2024-10-07T15:30:10.786238+01:00",
}

ProcedureCallError: This is a procedure error!
      at new ProcedureCallError (/user/robin/examples/simple/bindings.ts:301:5)
      at /user/robin/examples/simple/bindings.ts:228:15
```

> [!NOTE]
> This example is configured to throw on failure as you would prefer to if you are using it with something like React Query or Solid.js's `createResource`, you can disable this and get all responses back as the result type which can then be destructured to check or access the error or data.

When `ThrowOnError` is disabled, you get back a result type which can then further be narrowed to force error checks by enabling the `UseUnionResult` option which will only allow access to either the data or the error field depending on a guarded check of the `ok` field.

```text
todos ->  {
  ok: true,
  data: [
    {
      title: "Hello world!",
      completed: false,
      created_at: "2024-10-07T15:34:39.081796+01:00",
    }, {
      title: "Hello world again!",
      completed: true,
      created_at: "2024-10-07T15:34:39.081796+01:00",
    }
  ],
}
newTodo ->  {
  ok: true,
  data: {
    title: "Buy milk",
    completed: false,
    created_at: "2024-10-07T15:34:39.082366+01:00",
  },
}
t ->  {
  ok: false,
  error: "This is a procedure error!",
}
```

You can find this example presented here in the [`examples/simple`](./examples/simple) folder or a more application-like example [here](https://github.com/aosasona/robin-todo) using Solid.js, [BoltDB](https://github.com/etcd-io/bbolt) and Robin.

# Contributing

I cannot promise to review or merge contributions at the moment, at all in this state or speedily, but ideas (and perhaps even code) are always welcome!
