package math

import "testing"

func TestEvalMoneyExpression(t *testing.T) {
	cases := []struct {
		expr string
		want int64
	}{
		{"1200-10", 119000},
		{"1200%10", 12000},
		{"1200%12.5", 15000},
		{"7.5%2", 15},
		{"12.5*(3-1)/4", 625},
		{"10/3", 333},
		{"2+3*4", 1400},
		{"(2+3)*4", 2000},
	}

	for _, tc := range cases {
		got, err := EvalMoneyExpression(tc.expr)
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", tc.expr, err)
		}
		if got != tc.want {
			t.Fatalf("wrong result for %q: got %d want %d", tc.expr, got, tc.want)
		}
	}
}
