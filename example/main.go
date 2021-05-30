package main

import (
	"fmt"
	"goper"
)

const (
	Foo   = "foo"
	Bar   = "bar"
	Group = "group"
)

func main() {
	var g goper.Pool
	defer g.Shutdown()
	g.Default(Foo, 1, FooSay)
	g.Default(Bar, 1, BarSay)
	g.Groud(Group, 4)

	g.Put(Foo, "hello bar.")
	g.Put(Bar, "hello foo.")
	g.GroupPut(Group, GroupSay)
	g.Put(Group, GrouptFoo)

}

func FooSay(arg interface{}) {
	fmt.Println("foo say:", arg)

}

func BarSay(arg interface{}) {
	fmt.Println("bar say:", arg)
}

func GroupSay() {
	fmt.Println("Group:hello world.")
}

func GrouptFoo() {
	fmt.Println("Group:hello foo.")
}
