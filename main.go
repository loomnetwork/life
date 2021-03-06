package main

import (
	"flag"
	"fmt"
	"github.com/perlin-network/life/exec"
	"github.com/perlin-network/life/gowasm"
	"io/ioutil"
	"time"
)

func main() {
	entryFunctionFlag := flag.String("entry", "app_main", "entry function id")
	jitFlag := flag.Bool("jit", false, "enable jit")
	flag.Parse()

	// Read WebAssembly *.wasm file.
	input, err := ioutil.ReadFile(flag.Arg(0))
	if err != nil {
		panic(err)
	}

	// Instantiate a new WebAssembly VM with a few resolved imports.
	vm, err := exec.NewVirtualMachine(input, exec.VMConfig{
		EnableJIT:          *jitFlag,
		DefaultMemoryPages: 128,
		DefaultTableSize:   65536,
	}, gowasm.NewResolver(), nil)

	if err != nil {
		panic(err)
	}

	// Get the function ID of the entry function to be executed.
	entryID, ok := vm.GetFunctionExport(*entryFunctionFlag)
	if !ok {
		fmt.Printf("Entry function %s not found; starting from 0.\n", *entryFunctionFlag)
		entryID = 0
	}

	start := time.Now()

	// If any function prior to the entry function was declared to be
	// called by the module, run it first.
	if vm.Module.Base.Start != nil {
		startID := int(vm.Module.Base.Start.Index)
		_, err := vm.Run(startID)
		if err != nil {
			vm.PrintStackTrace()
			panic(err)
		}
	}

	// Run the WebAssembly module's entry function.
	ret, err := vm.Run(entryID, 0, 0)
	if err != nil {
		vm.PrintStackTrace()
		panic(err)
	}
	end := time.Now()

	fmt.Printf("return value = %d, duration = %v\n", ret, end.Sub(start))
}