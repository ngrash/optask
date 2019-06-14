package server

import "fmt"

func foo() {
	fmt.Println("foo")
}

func Moin() {
	foo()
	fmt.Println("Moin")
}
