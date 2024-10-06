package handler

import (
	"fmt"
	"time"

	"go.trulyao.dev/robin"
)

type User struct {
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (h *handler) Me(ctx *robin.Context, _ robin.Void) (User, error) {
	fmt.Println("Me")
	return User{}, nil
}

func (h *handler) SignIn(ctx *robin.Context, _ robin.Void) (*User, error) {
	return nil, nil
}

func (h *handler) SignUp(ctx *robin.Context, _ robin.Void) (*User, error) {
	return nil, nil
}
