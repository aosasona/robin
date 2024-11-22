package robin_test

import (
	"testing"

	"go.trulyao.dev/robin"
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
		{"matching name with multiple dots", "get_user", "lookup.user.profile", "lookup.user.profile"},
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
				q.SetAlias(test.alias)
			}

			if alias := q.Alias(); alias != test.expected {
				t.Errorf("expected %s, got %s", test.expected, alias)
			}
		})
	}
}
