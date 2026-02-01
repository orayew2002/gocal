package math

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"unicode"
)

type TokenType int

const (
	TNumber TokenType = iota
	TOp
	TLParen
	TRParen
)

type Token struct {
	Typ   TokenType
	Text  string
	Value float64
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

		if isNumStart(s, i) {
			start := i
			dotCount := 0

			for i < len(s) {
				c := s[i]
				if c == '.' {
					dotCount++
					if dotCount > 1 {
						return nil, fmt.Errorf("неверное число возле %q", s[start:i+1])
					}
					i++
					continue
				}
				if c >= '0' && c <= '9' {
					i++
					continue
				}
				break
			}

			txt := s[start:i]
			val, err := strconv.ParseFloat(txt, 64)
			if err != nil {
				return nil, fmt.Errorf("не удалось распарсить число %q: %w", txt, err)
			}

			tokens = append(tokens, Token{Typ: TNumber, Text: txt, Value: val})
			continue
		}

		return nil, fmt.Errorf("неожиданный символ: %q", string(s[i]))
	}

	return tokens, nil
}

func isOpByte(b byte) bool {
	return b == '+' || b == '-' || b == '*' || b == '/' || b == '^'
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
	case "^":
		return 3
	case "*", "/":
		return 2
	case "+", "-":
		return 1
	default:
		return 0
	}
}

func rightAssociative(op string) bool {
	return op == "^" || op == "NEG"
}

func toRPN(tokens []Token) ([]Token, error) {
	var out []Token
	var stack []Token
	var prev *Token

	for i := 0; i < len(tokens); i++ {
		t := tokens[i]

		switch t.Typ {
		case TNumber:
			out = append(out, t)

		case TLParen:
			stack = append(stack, t)

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
				return nil, errors.New("несогласованные скобки")
			}

		case TOp:
			op := t.Text
			if op == "-" && (prev == nil || prev.Typ == TOp || prev.Typ == TLParen) {
				op = "NEG"
				t.Text = "NEG"
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
			return nil, errors.New("неизвестный токен")
		}

		prev = &tokens[i]
	}

	for len(stack) > 0 {
		top := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if top.Typ == TLParen || top.Typ == TRParen {
			return nil, errors.New("несогласованные скобки")
		}
		out = append(out, top)
	}

	return out, nil
}

func evalRPN(rpn []Token) (float64, error) {
	var st []float64

	pop := func() (float64, error) {
		if len(st) == 0 {
			return 0, errors.New("не хватает операндов")
		}
		v := st[len(st)-1]
		st = st[:len(st)-1]
		return v, nil
	}

	for _, t := range rpn {
		switch t.Typ {
		case TNumber:
			st = append(st, t.Value)

		case TOp:
			switch t.Text {
			case "NEG":
				a, err := pop()
				if err != nil {
					return 0, err
				}
				st = append(st, -a)

			case "+", "-", "*", "/", "^":
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
				case "^":
					res = math.Pow(a, b)
				}
				st = append(st, res)

			default:
				return 0, fmt.Errorf("неизвестный оператор: %q", t.Text)
			}

		default:
			return 0, errors.New("неожиданный токен в RPN")
		}
	}

	if len(st) != 1 {
		return 0, errors.New("ошибка выражения: лишние значения")
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
