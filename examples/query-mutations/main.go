package main

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"go.trulyao.dev/robin"
)

const port = 8081

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
	ID   int    `json:"id,omitempty"`
	Name string `json:"name"`
}

var users = []User{
	{ID: 1, Name: "John Doe"},
	{ID: 2, Name: "Jane Doe"},
	{ID: 3, Name: "Alice"},
	{ID: 4, Name: "Bob"},
	{ID: 5, Name: "Charlie"},
}

type SerializableCustomError struct {
	Message string
	Code    int
}

func (s *SerializableCustomError) MarshalJSON() ([]byte, error) {
	// return []byte(fmt.Sprintf(`{"message":"%s","code":%d}`, s.Message, s.Code)), nil
	// OR
	// Remember that doing `json.Marshal(s)` will cause an infinite recursion
	return json.Marshal(map[string]interface{}{
		"message": s.Message,
		"code":    s.Code,
	})
}

func errorHandler(err error) (robin.Serializable, int) {
	message := err.Error()
	code := 500

	if e, ok := err.(Error); ok {
		code = e.Code
		message = e.Message
	} else if e, ok := err.(robin.Error); ok {
		code = e.Code
		message = "robin error: " + e.Message
	}

	return &SerializableCustomError{Message: message, Code: code}, code
}

func main() {
	r, err := robin.New(robin.Options{
		ErrorHandler:    errorHandler,
		EnableDebugMode: true,
	})
	if err != nil {
		slog.Error("Failed to create Robin instance", slog.String("error", err.Error()))
		return
	}

	instance, err := r.
		Add(robin.Query("ping", ping)).
		Add(robin.Query("getUser", getUser)).
		Add(robin.Query("getUsersByIds", getUsersByIds)).
		Add(robin.Query("getUsers", getUsers)).
		Add(robin.Query("error", errorEndpoint)).
		Add(robin.Mutation("addUser", addUser)).
		Add(robin.Mutation("deleteUser", deleteUser)).
		Add(robin.Mutation("error", errorEndpoint)).
		Build()
	if err != nil {
		slog.Error("Failed to build Robin instance", slog.String("error", err.Error()))
		return
	}

	instance.Serve()

	// Alternatively, you can use the default handler with your own mux and server
	//
	// mux := http.NewServeMux()
	// mux.Handle("POST /_robin", r.Handler())
	//
	// fmt.Printf("Server is running on port %d\n", port)
	// http.ListenAndServe(fmt.Sprintf(":%d", port), mux)

	// Or using the default handler, you will have to modify your endpoint to just `/` in the client side
	//
	// server := &http.Server{
	// 	Addr:    ":8080",
	// 	Handler: r.Handler(),
	// }
	//
	// server.ListenAndServe()
}

func ping(ctx *robin.Context, _ robin.Void) (string, error) {
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

func getUsers(ctx *robin.Context, _ robin.Void) ([]User, error) {
	return users, nil
}

func getUsersByIds(_ *robin.Context, ids []int) ([]User, error) {
	var foundUsers []User

	for _, id := range ids {
		for _, user := range users {
			if user.ID == id {
				foundUsers = append(foundUsers, user)
				break
			}
		}
	}

	return foundUsers, nil
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

func errorEndpoint(_ *robin.Context, _ robin.Void) (robin.Void, error) {
	return robin.Void{}, NewError("This is an error", 400)
}
