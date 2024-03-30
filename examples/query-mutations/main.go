package main

import (
	"fmt"
	"net/http"

	"go.trulyao.dev/robin"
)

type User struct {
	Name string
}

type Error struct {
	Message string
	Code    int
}

func (e *Error) Error() string {
	return e.Message
}

func NewError(message string, code int) *Error {
	return &Error{Message: message, Code: code}
}

var users = []User{
	{Name: "John Doe"},
}

func main() {
	errorHandler := func(err error) (any, int) {
		code := 500

		if e, ok := err.(*Error); ok {
			code = e.Code
		}

		return Error{Message: err.Error()}, code
	}

	r := robin.New(&robin.Options{ErrorHandler: errorHandler})

	r.Add(robin.Query("ping", ping)).
		Add(robin.Query("getUser", getUser)).
		Add(robin.Query("getUsers", getUsers)).
		Add(robin.Mutation("addUser", addUser)).
		Add(robin.Mutation("deleteUser", deleteUser))

	mux := http.NewServeMux()
	mux.Handle("/_robin", r.Handler())

	fmt.Println("Server is running on port 8080")
	http.ListenAndServe(":8080", mux)

	// Or using the default handler, you will have to modify your endpoint to just `/` in the client side
	//
	// server := &http.Server{
	// 	Addr:    ":8080",
	// 	Handler: r.Handler(),
	// }
	//
	// server.ListenAndServe()
}

func ping(ctx *robin.Context, _ string) (string, error) {
	return "pong", nil
}

func getUser(_ *robin.Context, name string) (User, error) {
	for _, user := range users {
		if user.Name == name {
			return user, nil
		}
	}
	return User{}, fmt.Errorf("user %s not found", name)
}

func getUsers(ctx *robin.Context, _ string) ([]User, error) {
	return users, nil
}

func addUser(_ *robin.Context, user User) (User, error) {
	users = append(users, user)
	return user, nil
}

func deleteUser(_ *robin.Context, name string) (User, error) {
	for i, user := range users {
		if user.Name == name {
			users = append(users[:i], users[i+1:]...)
			return user, nil
		}
	}
	return User{}, fmt.Errorf("user %s not found", name)
}
