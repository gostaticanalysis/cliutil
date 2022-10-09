package cliutil_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/gostaticanalysis/cliutil"
)

func TestSplit(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		want []any
	}{
		{"int", []any{"int", "", "", false}},
		{"*int", []any{"int", "", "", true}},
		{"testing.T", []any{"testing", "T", "", false}},
		{"*testing.T", []any{"testing", "T", "", true}},
		{"(*testing.T).Fatal", []any{"testing", "T", "Fatal", true}},
		{"(net/http.Handler).ServeHTTP", []any{"net/http", "Handler", "ServeHTTP", false}},
		{"(example.com/pkg.Type).Method", []any{"example.com/pkg", "Type", "Method", false}},
		{"(*example.com/pkg.Type).Method", []any{"example.com/pkg", "Type", "Method", true}},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			first, second, third, ptr := cliutil.Split(tt.name)
			got := []any{first, second, third, ptr}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}
