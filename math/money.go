package math

import (
	"errors"
	"fmt"
	"math"
	"math/bits"
	"strconv"
	"strings"
)

const moneyScale int64 = 100
const percentScale int64 = 10000

// EvalMoneyExpression evaluates an expression using fixed-point (cents) arithmetic.
// It returns the result in cents to avoid floating point errors.
func EvalMoneyExpression(expr string) (int64, error) {
	toks, err := tokenize(expr)
	if err != nil {
		return 0, err
	}
	rpn, err := toRPN(toks)
	if err != nil {
		return 0, err
	}
	return evalRPNMoney(rpn)
}

func evalRPNMoney(rpn []Token) (int64, error) {
	var st []int64

	pop := func() (int64, error) {
		if len(st) == 0 {
			return 0, errors.New("not enough operands")
		}
		v := st[len(st)-1]
		st = st[:len(st)-1]
		return v, nil
	}
	popN := func(n int) ([]int64, error) {
		if n < 0 {
			return nil, errors.New("invalid argument count")
		}
		if len(st) < n {
			return nil, errors.New("not enough operands")
		}
		vals := make([]int64, n)
		for i := n - 1; i >= 0; i-- {
			vals[i] = st[len(st)-1]
			st = st[:len(st)-1]
		}
		return vals, nil
	}

	for _, t := range rpn {
		switch t.Typ {
		case TNumber:
			if !isNumericLiteral(t.Text) {
				return 0, fmt.Errorf("non-numeric literal %q not supported in money expressions", t.Text)
			}
			v, err := parseCents(t.Text)
			if err != nil {
				return 0, err
			}
			st = append(st, v)

		case TFunc:
			switch t.Text {
			case "abs":
				if t.Arity != 1 {
					return 0, fmt.Errorf("function %q expects 1 argument", t.Text)
				}
				args, err := popN(1)
				if err != nil {
					return 0, err
				}
				if args[0] == math.MinInt64 {
					return 0, errors.New("overflow while computing abs")
				}
				if args[0] < 0 {
					st = append(st, -args[0])
				} else {
					st = append(st, args[0])
				}

			case "min", "max":
				if t.Arity < 2 {
					return 0, fmt.Errorf("function %q expects at least 2 arguments", t.Text)
				}
				args, err := popN(t.Arity)
				if err != nil {
					return 0, err
				}
				res := args[0]
				for i := 1; i < len(args); i++ {
					if t.Text == "min" {
						if args[i] < res {
							res = args[i]
						}
					} else {
						if args[i] > res {
							res = args[i]
						}
					}
				}
				st = append(st, res)

			default:
				return 0, fmt.Errorf("function %q not supported in money expressions", t.Text)
			}

		case TOp:
			switch t.Text {
			case "NEG":
				a, err := pop()
				if err != nil {
					return 0, err
				}
				if a == math.MinInt64 {
					return 0, errors.New("overflow while negating value")
				}
				st = append(st, -a)

			case "POS":
				a, err := pop()
				if err != nil {
					return 0, err
				}
				st = append(st, a)

			case "+", "-", "*", "/", "%":
				b, err := pop()
				if err != nil {
					return 0, err
				}
				a, err := pop()
				if err != nil {
					return 0, err
				}

				var res int64
				switch t.Text {
				case "+":
					res, err = addInt64(a, b)
				case "-":
					res, err = subInt64(a, b)
				case "*":
					var prod int64
					prod, err = mulInt64(a, b)
					if err == nil {
						res, err = divRound(prod, moneyScale)
					}
				case "/":
					var num int64
					num, err = mulInt64(a, moneyScale)
					if err == nil {
						res, err = divRound(num, b)
					}
				case "%":
					var prod int64
					prod, err = mulInt64(a, b)
					if err == nil {
						res, err = divRound(prod, percentScale)
					}
				}
				if err != nil {
					return 0, err
				}
				st = append(st, res)

			default:
				return 0, fmt.Errorf("operator %q not supported in money expressions", t.Text)
			}

		default:
			return 0, errors.New("unexpected token in RPN")
		}
	}

	if len(st) != 1 {
		return 0, errors.New("expression error: extra values")
	}
	return st[0], nil
}

func parseCents(txt string) (int64, error) {
	if txt == "" {
		return 0, errors.New("empty number")
	}
	if strings.ContainsAny(txt, "eE") {
		return 0, fmt.Errorf("exponent notation not supported in money expressions: %q", txt)
	}
	if strings.Count(txt, ".") > 1 {
		return 0, fmt.Errorf("invalid money number %q", txt)
	}

	parts := strings.SplitN(txt, ".", 2)
	intPart := parts[0]
	fracPart := ""
	if len(parts) == 2 {
		fracPart = parts[1]
	}
	if intPart == "" {
		intPart = "0"
	}
	if fracPart == "" {
		fracPart = "0"
	}
	if !allDigits(intPart) || !allDigits(fracPart) {
		return 0, fmt.Errorf("invalid money number %q", txt)
	}
	if len(fracPart) > 2 {
		return 0, fmt.Errorf("too many decimal places in %q", txt)
	}
	for len(fracPart) < 2 {
		fracPart += "0"
	}

	whole, err := strconv.ParseInt(intPart, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid money number %q: %w", txt, err)
	}
	frac, err := strconv.ParseInt(fracPart, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid money number %q: %w", txt, err)
	}

	wholeCents, err := mulInt64(whole, moneyScale)
	if err != nil {
		return 0, fmt.Errorf("money value overflow for %q", txt)
	}
	return addInt64(wholeCents, frac)
}

func isNumericLiteral(txt string) bool {
	if txt == "" {
		return false
	}
	if strings.ContainsAny(txt, "eE") {
		return false
	}
	dotSeen := false
	for i := 0; i < len(txt); i++ {
		c := txt[i]
		if c == '.' {
			if dotSeen {
				return false
			}
			dotSeen = true
			continue
		}
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func allDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func addInt64(a, b int64) (int64, error) {
	if b > 0 && a > math.MaxInt64-b {
		return 0, errors.New("overflow while adding values")
	}
	if b < 0 && a < math.MinInt64-b {
		return 0, errors.New("overflow while adding values")
	}
	return a + b, nil
}

func subInt64(a, b int64) (int64, error) {
	if b > 0 && a < math.MinInt64+b {
		return 0, errors.New("overflow while subtracting values")
	}
	if b < 0 && a > math.MaxInt64+b {
		return 0, errors.New("overflow while subtracting values")
	}
	return a - b, nil
}

func mulInt64(a, b int64) (int64, error) {
	if a == 0 || b == 0 {
		return 0, nil
	}
	signNegative := (a < 0) != (b < 0)
	ua := absUint64(a)
	ub := absUint64(b)
	hi, lo := bits.Mul64(ua, ub)
	if hi != 0 {
		return 0, errors.New("overflow while multiplying values")
	}
	if !signNegative {
		if lo > uint64(math.MaxInt64) {
			return 0, errors.New("overflow while multiplying values")
		}
		return int64(lo), nil
	}
	if lo > uint64(math.MaxInt64)+1 {
		return 0, errors.New("overflow while multiplying values")
	}
	if lo == uint64(math.MaxInt64)+1 {
		return math.MinInt64, nil
	}
	return -int64(lo), nil
}

func divRound(n, d int64) (int64, error) {
	if d == 0 {
		return 0, errors.New("division by zero")
	}
	q := n / d
	r := n % d
	if r == 0 {
		return q, nil
	}

	absR := absUint64(r)
	absD := absUint64(d)
	if absR*2 >= absD {
		if (n > 0) == (d > 0) {
			q++
		} else {
			q--
		}
	}
	return q, nil
}

func absUint64(v int64) uint64 {
	if v >= 0 {
		return uint64(v)
	}
	return uint64(^v) + 1
}
