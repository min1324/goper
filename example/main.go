package main

import (
	"fmt"
	"goper"
)

func main() {
	var g goper.Goper
	g.Default(1, Router)
	defer g.Close()

	g.Deliver("hello world.")
}

func Router(i interface{}) {
	s, ok := i.(string)
	if ok {
		fmt.Println(s)
	}
}
