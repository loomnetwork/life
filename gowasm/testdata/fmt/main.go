package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("hello world")
	a := 100
	b := 1000
	c := b - 2*a
	fmt.Printf("%d - %d = %d\n", b, 2*a, c)
	os.Exit(10)
}
