package handler

import (
	"net/http"

	"todo/repository"

	apperrors "todo/pkg/errors"

	"go.trulyao.dev/robin"
)

type SignInInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *handler) Me(ctx *robin.Context, _ robin.Void) (string, error) {
	return "me", nil
}

func (h *handler) SignIn(ctx *robin.Context, data SignInInput) (repository.User, error) {
	user, err := h.repository.UserRepo().FindByUsername(data.Username)
	if err != nil {
		return repository.User{}, err
	}

	if !user.VerifyPassword(data.Password) {
		return repository.User{}, apperrors.New(http.StatusUnauthorized, "Invalid credentials")
	}

	return user, nil
}

func (h *handler) SignUp(
	ctx *robin.Context,
	data repository.CreateUserInput,
) (repository.User, error) {
	if data.Username == "" {
		return repository.User{}, apperrors.New(http.StatusBadRequest, "Username is required")
	}

	if data.Password == "" {
		return repository.User{}, apperrors.New(http.StatusBadRequest, "Password is required")
	}

	if len(data.Password) < 6 {
		return repository.User{}, apperrors.New(
			http.StatusBadRequest,
			"Password must be at least 6 characters long",
		)
	}

	user, err := h.repository.UserRepo().Create(data)
	if err != nil {
		return repository.User{}, err
	}

	return user, nil
}
