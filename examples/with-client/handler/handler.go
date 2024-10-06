package handler

import (
	"encoding/base64"
	"log/slog"
	"net/http"

	"todo/repository"

	apperrors "todo/pkg/errors"

	"go.trulyao.dev/robin"
)

type handler struct {
	repository repository.Repository
}

type CreateInput struct {
	Title string `json:"title"`
}

func New(repo repository.Repository) *handler {
	return &handler{repository: repo}
}

func (h *handler) RequireAuth(c *robin.Context) error {
	authCookie, err := c.Request().Cookie("auth")
	if err != nil {
		return apperrors.New(http.StatusUnauthorized, "Unauthorized")
	}

	username, err := base64.StdEncoding.DecodeString(authCookie.Value)
	if err != nil {
		slog.Error("Failed to decode auth cookie", slog.String("cookie", authCookie.Value))
		return apperrors.New(http.StatusUnauthorized, "Unauthorized")
	}

	user, err := h.repository.UserRepo().FindByUsername(string(username))
	if err != nil {
		slog.Error("Failed to find user", slog.String("username", string(username)))
		return apperrors.New(http.StatusUnauthorized, "Unauthorized")
	}

	c.Set("user", user)
	return nil
}
