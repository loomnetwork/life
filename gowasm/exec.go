package gowasm

import (
	"fmt"
	"github.com/perlin-network/life/exec"
	"io/ioutil"
)

func RunWASMFile(path string, entry string) (ret int64, err error) {
	return RunWASMFileWithResolver(NewResolver(), path, entry)
}

func RunWASMFileWithResolver(resolver *Resolver, path string, entry string) (ret int64, err error) {
	if len(entry) < 1 {
		entry = "run"
	}

	// Read WebAssembly *.wasm file.
	input, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	// Instantiate a new WebAssembly VM with a few resolved imports.
	vm, err := exec.NewVirtualMachine(input, exec.VMConfig{
		EnableJIT:          false,
		DefaultMemoryPages: 128,
		DefaultTableSize:   65536,
	}, resolver, nil)

	if err != nil {
		return
	}

	// Get the function ID of the entry function to be executed.
	entryID, ok := vm.GetFunctionExport(entry)
	if !ok {
		fmt.Printf("Entry function %s not found; starting from 0.\n", entry)
		entryID = 0
	}

	// If any function prior to the entry function was declared to be
	// called by the module, run it first.
	if vm.Module.Base.Start != nil {
		startID := int(vm.Module.Base.Start.Index)
		_, err = vm.Run(startID)
		if err != nil {
			vm.PrintStackTrace()
			return
		}
	}

	// Run the WebAssembly module's entry function.
	ret, err = vm.Run(entryID, 0, 0)
	if err != nil {
		vm.PrintStackTrace()
		return
	}

	return
}