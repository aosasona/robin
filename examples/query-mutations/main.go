package main

import (
	"fmt"
	"net/http"

	"go.trulyao.dev/robin"
)

type Error struct {
	Message string
	Code    int
}

func (e Error) MarshalJSON() ([]byte, error) {
	return []byte(e.Message), nil
}

func (e Error) Error() string {
	return e.Message
}

func NewError(message string, code int) *Error {
	return &Error{Message: message, Code: code}
}

type User struct {
	ID   int
	Name string
}

var users = []User{
	{ID: 1, Name: "John Doe"},
	{ID: 2, Name: "Jane Doe"},
	{ID: 3, Name: "Alice"},
	{ID: 4, Name: "Bob"},
	{ID: 5, Name: "Charlie"},
}

func main() {
	errorHandler := func(err error) ([]byte, int) {
		message := err.Error()
		code := 500

		if e, ok := err.(Error); ok {
			code = e.Code
			message = e.Message
		} else if e, ok := err.(robin.Error); ok {
			code = e.Code
			message = "robin error: " + e.Message
		}

		return []byte("[via custom handler] " + message), code
	}

	r := robin.New(&robin.Options{ErrorHandler: errorHandler, EnableDebugMode: true})

	r.Add(robin.Query("ping", ping)).
		Add(robin.Query("getUser", getUser)).
		Add(robin.Query("getUsers", getUsers)).
		Add(robin.Mutation("addUser", addUser)).
		Add(robin.Mutation("deleteUser", deleteUser))

	mux := http.NewServeMux()
	mux.Handle("POST /_robin", r.Handler())

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

func getUser(_ *robin.Context, id int) (User, error) {
	if id == 0 {
		return User{}, Error{Message: "User ID is required", Code: 400}
	}

	for _, user := range users {
		if user.ID == id {
			return user, nil
		}
	}
	return User{}, Error{Message: "User not found", Code: 400}
}

func getUsers(ctx *robin.Context, _ string) ([]User, error) {
	return users, nil
}

func addUser(_ *robin.Context, user User) (User, error) {
	if user.Name == "" {
		return User{}, Error{Message: "User name is required", Code: 400}
	}

	user.ID = len(users) + 1
	users = append(users, user)

	return user, nil
}

func deleteUser(_ *robin.Context, id int) (User, error) {
	if id == 0 {
		return User{}, Error{Message: "User ID is required", Code: 400}
	}

	for i, user := range users {
		if user.ID == id {
			users = append(users[:i], users[i+1:]...)
			return user, nil
		}
	}

	return User{}, Error{Message: fmt.Sprintf("User %d not found", id), Code: 400}
}
