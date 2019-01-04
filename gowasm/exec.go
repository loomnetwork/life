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
	vm := NewVirtualMachine(resolver, path, entry)
	err = vm.Init()
	if err != nil {
		return
	}
	return vm.Run()
}

type VirtualMachine struct {
	resolver *Resolver
	filePath string
	entry    string

	vm *exec.VirtualMachine
}

func NewVirtualMachine(resolver *Resolver, filePath string, entry string) *VirtualMachine {
	return &VirtualMachine{
		resolver: resolver,
		filePath: filePath,
		entry:    entry,
	}
}

func (vm *VirtualMachine) Init() (err error) {
	if len(vm.entry) < 1 {
		vm.entry = "run"
	}

	// Read WebAssembly *.wasm file.
	input, err := ioutil.ReadFile(vm.filePath)
	if err != nil {
		return
	}

	// Instantiate a new WebAssembly VM with a few resolved imports.
	vm.vm, err = exec.NewVirtualMachine(input, exec.VMConfig{
		EnableJIT:          false,
		DefaultMemoryPages: 128,
		DefaultTableSize:   65536,
	}, vm.resolver, nil)

	return
}

func (vm *VirtualMachine) Run() (ret int64, err error) {
	ret, err = vm.run(vm.vm)
	vm.vm.Reset()
	return
}

func (vm *VirtualMachine) RunInCloneVM() (ret int64, err error) {
	m, err := vm.vm.Clone()
	if err != nil {
		return
	}

	return vm.run(m)
}

func (vm *VirtualMachine) run(m *exec.VirtualMachine) (ret int64, err error) {
	// Get the function ID of the entry function to be executed.
	entryID, ok := m.GetFunctionExport(vm.entry)
	if !ok {
		fmt.Printf("Entry function %s not found; starting from 0.\n", vm.entry)
		entryID = 0
	}

	// If any function prior to the entry function was declared to be
	// called by the module, run it first.
	if m.Module.Base.Start != nil {
		startID := int(m.Module.Base.Start.Index)
		_, err = m.Run(startID)
		if err != nil {
			m.PrintStackTrace()
			return
		}
	}

	// Run the WebAssembly module's entry function.
	ret, err = m.Run(entryID)
	if err != nil {
		m.PrintStackTrace()
		return
	}

	return
}
