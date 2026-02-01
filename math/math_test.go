package math

import (
	"math"
	"testing"
)

func TestEvalExpression_AllOperators(t *testing.T) {
	cases := []struct {
		expr string
		want float64
	}{
		{"12.5*(3-1)/4", 6.25},
		{"2+3*4", 14},
		{"10-6/3", 8},
		{"(-3)+5", 2},
		{"2*-3", -6},
		{"2^3^2", 512},
		{"5%2", 1},
		{"7.5%2", 1.5},
		{"(2+3)^(1+1)", 25},
		{"-(3+4)*2", -14},
		{"2^-3", 0.125},
		{"1.5e2+2.5e-1", 150.25},
	}

	for _, tc := range cases {
		got, err := EvalExpression(tc.expr)
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", tc.expr, err)
		}
		if math.Abs(got-tc.want) > 1e-9 {
			t.Fatalf("wrong result for %q: got %v want %v", tc.expr, got, tc.want)
		}
	}
}

func TestEvalExpression_Advanced(t *testing.T) {
	cases := []struct {
		expr string
		want float64
	}{
		{"sin(pi/2)+cos(0)", 2},
		{"sqrt(2)^2", math.Pow(math.Sqrt(2), 2)},
		{"ln(e^2)", math.Log(math.Pow(math.E, 2))},
		{"log(1000)", math.Log10(1000)},
		{"exp(1)", math.Exp(1)},
		{"min(5,2,7,3)", 2},
		{"max(5,2,7,3)", 7},
		{"floor(3.9)+ceil(2.1)", 6},
		{"round(2.6)+round(2.4)", 5},
		{"abs(-3.5)+abs(2.25)", 5.75},
		{"max(10, 6%4 + 2^3) - min(5, 3+1)", 6},
		{"sqrt(81)+5*(2+3)^2-10/5", 132},
		{"sin(pi/6)^2+cos(pi/6)^2", math.Pow(math.Sin(math.Pi/6), 2) + math.Pow(math.Cos(math.Pi/6), 2)},
		{"min(2, max(3, 4, 1))", 2},
		{"(-2.5)^3 + 10%3", -14.625},
		{"pow(2, 10) + atan2(1, 1)", math.Pow(2, 10) + math.Atan2(1, 1)},
		{"logn(8, 2) + log(100)", math.Log(8)/math.Log(2) + math.Log10(100)},
	}

	for _, tc := range cases {
		got, err := EvalExpression(tc.expr)
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", tc.expr, err)
		}
		if math.Abs(got-tc.want) > 1e-9 {
			t.Fatalf("wrong result for %q: got %v want %v", tc.expr, got, tc.want)
		}
	}
}
