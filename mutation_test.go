package robin_test

import (
	"io"
	"reflect"
	"testing"

	"go.trulyao.dev/robin"
	"go.trulyao.dev/robin/types"
)

func Test_MutationAlias(t *testing.T) {
	tests := []struct {
		description string
		name        string
		alias       string
		expected    string
	}{
		{"matching name without any alias", "create_user", "", "user"},
		{"NOT matching name without any alias", "my_user", "", "my.user"},
		{"matching name with alias", "create_user", "new-user", "new-user"},
		{"NOT matching name with alias", "del_user", "delete-user", "delete-user"},
		{"matching name with alias and dot", "create_user", "new.user", "new.user"},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			m := robin.M(test.name, func(ctx *robin.Context, body string) (string, error) {
				return "", nil
			})

			if test.alias != "" {
				m.WithAlias(test.alias)
			}

			if alias := m.Alias(); alias != test.expected {
				t.Errorf("expected %s, got %s", test.expected, alias)
			}
		})
	}
}

func Test_MutationWithRawPayloadPanic(t *testing.T) {
	type User struct {
		ID int `json:"id"`
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected a panic, but there was none")
		}
	}()

	// This should panic
	_ = robin.M("user.create", func(ctx *robin.Context, body string) (string, error) {
		return "", nil
	}).WithRawPayload(User{})
}

func Test_MutationWithRawPayload(t *testing.T) {
	type User struct {
		ID int `json:"id"`
	}

	mutationFn := func(ctx *robin.Context, body io.ReadCloser) (string, error) {
		return "", nil
	}

	m := robin.M("user.create", mutationFn).WithRawPayload(User{})

	// Ensure we get the valid payload type
	if m.ExpectedPayloadType() != types.ExpectedPayloadRaw {
		t.Errorf("expected %v, got %v", types.ExpectedPayloadRaw, m.ExpectedPayloadType())
	}

	// Ensure the overriden payload type is correct
	userType := reflect.TypeOf(User{})
	if payload := m.PayloadInterface(); reflect.TypeOf(payload).String() != userType.String() {
		t.Errorf("expected %v, got %v", userType, reflect.TypeOf(payload))
	}
}
