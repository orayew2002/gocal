package math

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
)

type TokenType int

const (
	TNumber TokenType = iota
	TOp
	TFunc
	TComma
	TLParen
	TRParen
)

type Token struct {
	Typ   TokenType
	Text  string
	Value float64
	Arity int
}

func tokenize(s string) ([]Token, error) {
	var tokens []Token
	i := 0

	for i < len(s) {
		r := rune(s[i])

		if unicode.IsSpace(r) {
			i++
			continue
		}

		if s[i] == ',' {
			tokens = append(tokens, Token{Typ: TComma, Text: ","})
			i++
			continue
		}
		if s[i] == '(' {
			tokens = append(tokens, Token{Typ: TLParen, Text: "("})
			i++
			continue
		}
		if s[i] == ')' {
			tokens = append(tokens, Token{Typ: TRParen, Text: ")"})
			i++
			continue
		}

		if isOpByte(s[i]) {
			tokens = append(tokens, Token{Typ: TOp, Text: string(s[i])})
			i++
			continue
		}

		if isIdentStart(s[i]) {
			start := i
			i++
			for i < len(s) && isIdentContinue(s[i]) {
				i++
			}
			name := strings.ToLower(s[start:i])
			if val, ok := constants[name]; ok {
				tokens = append(tokens, Token{Typ: TNumber, Text: name, Value: val})
			} else {
				tokens = append(tokens, Token{Typ: TFunc, Text: name})
			}
			continue
		}

		if isNumStart(s, i) {
			start := i
			dotCount := 0
			hasDigits := false

			for i < len(s) {
				c := s[i]
				if c == '.' {
					dotCount++
					if dotCount > 1 {
						return nil, fmt.Errorf("invalid number near %q", s[start:i+1])
					}
					i++
					continue
				}
				if c >= '0' && c <= '9' {
					hasDigits = true
					i++
					continue
				}
				if (c == 'e' || c == 'E') && hasDigits {
					i++
					if i < len(s) && (s[i] == '+' || s[i] == '-') {
						i++
					}
					expStart := i
					for i < len(s) && s[i] >= '0' && s[i] <= '9' {
						i++
					}
					if expStart == i {
						return nil, fmt.Errorf("invalid exponent in number near %q", s[start:i])
					}
					break
				}
				break
			}

			txt := s[start:i]
			val, err := strconv.ParseFloat(txt, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse number %q: %w", txt, err)
			}

			tokens = append(tokens, Token{Typ: TNumber, Text: txt, Value: val})
			continue
		}

		return nil, fmt.Errorf("unexpected character: %q", string(s[i]))
	}

	return tokens, nil
}

func isOpByte(b byte) bool {
	return b == '+' || b == '-' || b == '*' || b == '/' || b == '^' || b == '%'
}

func isIdentStart(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '_'
}

func isIdentContinue(b byte) bool {
	return isIdentStart(b) || (b >= '0' && b <= '9')
}

func isNumStart(s string, i int) bool {
	if i >= len(s) {
		return false
	}
	if s[i] >= '0' && s[i] <= '9' {
		return true
	}
	if s[i] == '.' && i+1 < len(s) && s[i+1] >= '0' && s[i+1] <= '9' {
		return true
	}
	return false
}

func precedence(op string) int {
	switch op {
	case "NEG":
		return 4
	case "POS":
		return 4
	case "^":
		return 3
	case "*", "/", "%":
		return 2
	case "+", "-":
		return 1
	default:
		return 0
	}
}

func rightAssociative(op string) bool {
	return op == "^" || op == "NEG" || op == "POS"
}

func toRPN(tokens []Token) ([]Token, error) {
	var out []Token
	var stack []Token
	var prev *Token
	var funcParen []bool
	var argCount []int

	for i := 0; i < len(tokens); i++ {
		t := tokens[i]

		switch t.Typ {
		case TNumber:
			out = append(out, t)

		case TFunc:
			if i+1 >= len(tokens) || tokens[i+1].Typ != TLParen {
				return nil, fmt.Errorf("function %q must be called with parentheses", t.Text)
			}
			stack = append(stack, t)

		case TLParen:
			stack = append(stack, t)
			if prev != nil && prev.Typ == TFunc {
				funcParen = append(funcParen, true)
				argCount = append(argCount, 0)
			} else {
				funcParen = append(funcParen, false)
				argCount = append(argCount, 0)
			}

		case TComma:
			found := false
			for len(stack) > 0 {
				top := stack[len(stack)-1]
				if top.Typ == TLParen {
					found = true
					break
				}
				stack = stack[:len(stack)-1]
				out = append(out, top)
			}
			if !found || len(funcParen) == 0 || !funcParen[len(funcParen)-1] {
				return nil, errors.New("comma must appear inside function arguments")
			}
			argCount[len(argCount)-1]++

		case TRParen:
			found := false
			for len(stack) > 0 {
				top := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				if top.Typ == TLParen {
					found = true
					break
				}
				out = append(out, top)
			}
			if !found {
				return nil, errors.New("mismatched parentheses")
			}
			if len(funcParen) == 0 {
				return nil, errors.New("mismatched parentheses")
			}
			isFuncCall := funcParen[len(funcParen)-1]
			argc := argCount[len(argCount)-1]
			funcParen = funcParen[:len(funcParen)-1]
			argCount = argCount[:len(argCount)-1]

			if isFuncCall {
				if prev != nil && prev.Typ == TLParen {
					argc = 0
				} else {
					argc = argc + 1
				}
				if len(stack) == 0 || stack[len(stack)-1].Typ != TFunc {
					return nil, errors.New("function call missing name")
				}
				fn := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				fn.Arity = argc
				out = append(out, fn)
			}

		case TOp:
			op := t.Text
			if (op == "-" || op == "+") && (prev == nil || prev.Typ == TOp || prev.Typ == TLParen || prev.Typ == TComma) {
				if op == "-" {
					op = "NEG"
				} else {
					op = "POS"
				}
				t.Text = op
			}

			for len(stack) > 0 {
				top := stack[len(stack)-1]
				if top.Typ != TOp {
					break
				}

				p1 := precedence(t.Text)
				p2 := precedence(top.Text)

				if (rightAssociative(t.Text) && p1 < p2) ||
					(!rightAssociative(t.Text) && p1 <= p2) {
					stack = stack[:len(stack)-1]
					out = append(out, top)
					continue
				}
				break
			}

			stack = append(stack, t)

		default:
			return nil, errors.New("unknown token")
		}

		prev = &tokens[i]
	}

	for len(stack) > 0 {
		top := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if top.Typ == TLParen || top.Typ == TRParen {
			return nil, errors.New("mismatched parentheses")
		}
		if top.Typ == TFunc {
			return nil, errors.New("function call missing parentheses")
		}
		out = append(out, top)
	}

	return out, nil
}

func evalRPN(rpn []Token) (float64, error) {
	var st []float64

	pop := func() (float64, error) {
		if len(st) == 0 {
			return 0, errors.New("not enough operands")
		}
		v := st[len(st)-1]
		st = st[:len(st)-1]
		return v, nil
	}
	popN := func(n int) ([]float64, error) {
		if n < 0 {
			return nil, errors.New("invalid argument count")
		}
		if len(st) < n {
			return nil, errors.New("not enough operands")
		}
		vals := make([]float64, n)
		for i := n - 1; i >= 0; i-- {
			vals[i] = st[len(st)-1]
			st = st[:len(st)-1]
		}
		return vals, nil
	}

	for _, t := range rpn {
		switch t.Typ {
		case TNumber:
			st = append(st, t.Value)

		case TFunc:
			switch t.Text {
			case "sin", "cos", "tan", "asin", "acos", "atan", "sqrt", "abs", "ln", "log", "exp", "floor", "ceil", "round":
				if t.Arity != 1 {
					return 0, fmt.Errorf("function %q expects 1 argument", t.Text)
				}
				args, err := popN(1)
				if err != nil {
					return 0, err
				}
				var res float64
				switch t.Text {
				case "sin":
					res = math.Sin(args[0])
				case "cos":
					res = math.Cos(args[0])
				case "tan":
					res = math.Tan(args[0])
				case "asin":
					res = math.Asin(args[0])
				case "acos":
					res = math.Acos(args[0])
				case "atan":
					res = math.Atan(args[0])
				case "sqrt":
					res = math.Sqrt(args[0])
				case "abs":
					res = math.Abs(args[0])
				case "ln":
					res = math.Log(args[0])
				case "log":
					res = math.Log10(args[0])
				case "exp":
					res = math.Exp(args[0])
				case "floor":
					res = math.Floor(args[0])
				case "ceil":
					res = math.Ceil(args[0])
				case "round":
					res = math.Round(args[0])
				}
				st = append(st, res)

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

			case "pow", "atan2":
				if t.Arity != 2 {
					return 0, fmt.Errorf("function %q expects 2 arguments", t.Text)
				}
				args, err := popN(2)
				if err != nil {
					return 0, err
				}
				if t.Text == "pow" {
					st = append(st, math.Pow(args[0], args[1]))
				} else {
					st = append(st, math.Atan2(args[0], args[1]))
				}

			case "logn":
				if t.Arity != 2 {
					return 0, fmt.Errorf("function %q expects 2 arguments", t.Text)
				}
				args, err := popN(2)
				if err != nil {
					return 0, err
				}
				st = append(st, math.Log(args[0])/math.Log(args[1]))

			default:
				return 0, fmt.Errorf("unknown function: %q", t.Text)
			}

		case TOp:
			switch t.Text {
			case "NEG":
				a, err := pop()
				if err != nil {
					return 0, err
				}
				st = append(st, -a)

			case "POS":
				a, err := pop()
				if err != nil {
					return 0, err
				}
				st = append(st, a)

			case "+", "-", "*", "/", "%", "^":
				b, err := pop()
				if err != nil {
					return 0, err
				}
				a, err := pop()
				if err != nil {
					return 0, err
				}

				var res float64
				switch t.Text {
				case "+":
					res = a + b
				case "-":
					res = a - b
				case "*":
					res = a * b
				case "/":
					res = a / b
				case "%":
					res = a * b / 100
				case "^":
					res = math.Pow(a, b)
				}
				st = append(st, res)

			default:
				return 0, fmt.Errorf("unknown operator: %q", t.Text)
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

func EvalExpression(expr string) (float64, error) {
	toks, err := tokenize(expr)
	if err != nil {
		return 0, err
	}
	rpn, err := toRPN(toks)
	if err != nil {
		return 0, err
	}
	return evalRPN(rpn)
}

var constants = map[string]float64{
	"pi": math.Pi,
	"e":  math.E,
}
