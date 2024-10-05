package generator_test

import (
	"testing"

	"go.trulyao.dev/robin/generator"
)

func Test_TestNormalizeName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"foo.bar", "fooBar"},
		{"foo.bar.baz", "fooBarBaz"},
		{"foo.bar-baz", "fooBarBaz"},
		{"foo.bar_baz", "fooBarBaz"},
		{"foo-bar", "fooBar"},
		{"foo_bar", "fooBar"},
		{"foo", "foo"},
		{"foo_bar_baz", "fooBarBaz"},
		{"foo-bar-baz", "fooBarBaz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generator.NormalizeProcedureName(tt.name); got != tt.want {
				t.Errorf("NormalizeProcedureName() = %v, want %v", got, tt.want)
			}
		})
	}
}
