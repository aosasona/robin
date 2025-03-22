package robin_test

import (
	"io"
	"reflect"
	"testing"

	"go.trulyao.dev/robin"
	"go.trulyao.dev/robin/types"
)

func Test_QueryAlias(t *testing.T) {
	tests := []struct {
		description string
		name        string
		alias       string
		expected    string
	}{
		{"matching name without any alias", "get_user", "", "user"},
		{"NOT matching name without any alias", "my_user", "", "my.user"},
		{"matching name with alias", "get_user", "lookup-user", "lookup-user"},
		{"NOT matching name with alias", "my_user", "lookup-user", "lookup-user"},
		{"matching name with alias and dot", "get_user", "lookup.user", "lookup.user"},
		{"NOT matching name with alias and dot", "my_user", "lookup.user", "lookup.user"},
		{
			"matching name with multiple dots",
			"get_user",
			"lookup.user.profile",
			"lookup.user.profile",
		},
		{"matching name with [lookup] prefix", "lookup_user", "", "user"},
		{"matching name with [fetch] prefix", "fetch_user", "", "user"},
		{"matching name with [get] prefix", "get_user", "", "user"},
		{"matching name with [query] prefix", "query_user", "", "user"},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			q := robin.Q(test.name, func(ctx *robin.Context, body string) (string, error) {
				return "", nil
			})

			if test.alias != "" {
				q.WithAlias(test.alias)
			}

			if alias := q.Alias(); alias != test.expected {
				t.Errorf("expected %s, got %s", test.expected, alias)
			}
		})
	}
}

func Test_QueryWithRawPayloadPanic(t *testing.T) {
	type User struct {
		ID int `json:"id"`
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected a panic, but there was none")
		}
	}()

	// This should panic
	_ = robin.Q("get_user", func(ctx *robin.Context, body string) (string, error) {
		return "", nil
	}).WithRawPayload(User{})
}

func Test_QueryWithRawPayload(t *testing.T) {
	type User struct {
		ID int `json:"id"`
	}

	q := robin.Q("user.find", func(ctx *robin.Context, body io.ReadCloser) (string, error) {
		return "", nil
	}).WithRawPayload(User{})

	// Ensure we get the valid payload type
	if q.ExpectedPayloadType() != types.ExpectedPayloadRaw {
		t.Errorf("expected %v, got %v", types.ExpectedPayloadRaw, q.ExpectedPayloadType())
	}

	// Ensure the overriden payload type is correct
	userType := reflect.TypeOf(User{})
	if payload := q.PayloadInterface(); reflect.TypeOf(payload).String() != userType.String() {
		t.Errorf("expected %v, got %v", userType, reflect.TypeOf(payload))
	}
}
