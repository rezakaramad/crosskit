package pdns

import (
	"testing"
)

func TestEnsureTrailingDot(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"rezakara.demo", "rezakara.demo."},
		{"rezakara.demo.", "rezakara.demo."},
		{"", "."},
	}

	for _, tc := range cases {
		got := EnsureTrailingDot(tc.input)
		if got != tc.want {
			t.Errorf("EnsureTrailingDot(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
