package main

import (
	"fmt"

	"github.com/orayew2002/gocal/math"
)

func main() {
	tests := []string{
		"12.5*(3-1)/4",
		"2+3*4",
		"(-3)+5",
		"2*-3",
		"2^3^2",
	}
	for _, s := range tests {
		v, err := math.EvalExpression(s)
		if err != nil {
			fmt.Println(err.Error())
		}

		fmt.Println(s, "=", v)
	}
}
