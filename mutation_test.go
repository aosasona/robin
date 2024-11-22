package robin_test

import (
	"testing"

	"go.trulyao.dev/robin"
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
