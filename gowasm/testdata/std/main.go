package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	fmt.Println("reading stdin")
	reader := bufio.NewReader(os.Stdin)
	line, isPrefix, err := reader.ReadLine()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s isPrefix %t\n", line, isPrefix)
}


