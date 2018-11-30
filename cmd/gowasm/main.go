package main

import (
	"flag"
	"github.com/perlin-network/life/gowasm"
	"os"
)

func main() {
	entryFunctionFlag := flag.String("entry", "run", "entry function id")
	flag.Parse()

	ret, err := gowasm.RunWASMFile(flag.Arg(0), *entryFunctionFlag)
	if err != nil {
		panic(err)
	}

	os.Exit(int(ret))
}
