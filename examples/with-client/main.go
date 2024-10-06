package main

// This demo does not represent a production-ready application nor best practices, this is simply for demonstration purposes.
import (
	"log"

	"todo/handler"
	"todo/repository"

	apperrors "todo/pkg/errors"

	"go.etcd.io/bbolt"
	"go.trulyao.dev/robin"
	"go.trulyao.dev/robin/types"
	_ "go.trulyao.dev/seer"
)

func initDB() *bbolt.DB {
	db, err := bbolt.Open("todos.db", 0o600, nil)
	if err != nil {
		log.Fatalf("Failed to open BoltDB: %s", err)
	}

	// Create buckets
	err = db.Update(func(tx *bbolt.Tx) error {
		if _, err = tx.CreateBucketIfNotExists([]byte("users")); err != nil {
			return err
		}

		if _, err = tx.CreateBucketIfNotExists([]byte("todos")); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Fatalf("Failed to create buckets: %s", err)
	}

	return db
}

func main() {
	r, err := robin.New(robin.Options{
		CodegenOptions: robin.CodegenOptions{
			Path:             "./ui/src/lib",
			GenerateBindings: true,
		},
		EnableDebugMode: false,
		ErrorHandler:    apperrors.ErrorHandler,
	})
	if err != nil {
		log.Fatalf("Failed to create a new Robin instance: %s", err)
	}

	db := initDB()
	repo := repository.New(db)
	h := handler.New(repo)

	i, err := r.
		// Queries
		Add(q("whoami", h.Me, h.RequireAuth)).
		Add(q("list-todos", h.List, h.RequireAuth)).
		// Mutations
		Add(m("sign-in", h.SignIn)).
		Add(m("sign-up", h.SignUp)).
		Add(m("create-todo", h.Create, h.RequireAuth)).
		Build()
	if err != nil {
		log.Fatalf("Failed to build Robin instance: %s", err)
	}

	if err := i.Export(); err != nil {
		log.Fatalf("Failed to export client: %s", err)
	}

	if err := i.Serve(); err != nil {
		log.Fatalf("Failed to serve Robin instance: %s", err)
		return
	}
}

func q[T, K any](
	name string,
	handler robin.QueryFn[T, K],
	middlewares ...types.Middleware,
) robin.Procedure {
	return robin.Query(name, handler).WithMiddleware(middlewares...)
}

func m[T, K any](
	name string,
	handler robin.MutationFn[T, K],
	middlewares ...types.Middleware,
) robin.Procedure {
	return robin.Mutation(name, handler).WithMiddleware(middlewares...)
}
