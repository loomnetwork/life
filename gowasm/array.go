package gowasm

import (
	"github.com/perlin-network/life/exec"
	"unsafe"
)

type jsInt8Array []int8

func (a jsInt8Array) New(vm *exec.VirtualMachine, args ...Value) interface{} {
	if len(args) != 3 {
		panic("jsInt8Array New len(args) != 3")
	}

	m := args[0]
	s := args[1]
	n := args[2]

	if s.IsNaNHead() || n.IsNaNHead() {
		return make(jsInt8Array, 0)
	}

	if m.ref == memory.ref {
		us := vm.Memory[s.Int() : s.Int()+n.Int()]
		r := *((*jsInt8Array)(unsafe.Pointer(&us)))
		return r

	}
	panic("jsInt8Array New invalid args")
}

type jsInt16Array []int16

func (a jsInt16Array) New(vm *exec.VirtualMachine, args ...Value) interface{} {
	if len(args) != 3 {
		panic("jsInt16Array New len(args) != 3")
	}

	m := args[0]
	s := args[1]
	n := args[2]

	if s.IsNaNHead() || n.IsNaNHead() {
		return make(jsInt16Array, 0)
	}

	if m.ref == memory.ref {
		us := vm.Memory[s.Int() : s.Int()+n.Int()*2]
		r := *((*jsInt16Array)(unsafe.Pointer(&us)))
		return r
	}
	panic("jsInt16Array New invalid args")
}

type jsInt32Array []int32

func (a jsInt32Array) New(vm *exec.VirtualMachine, args ...Value) interface{} {
	if len(args) != 3 {
		panic("jsInt32Array New len(args) != 3")
	}

	m := args[0]
	s := args[1]
	n := args[2]

	if s.IsNaNHead() || n.IsNaNHead() {
		return make(jsInt32Array, 0)
	}

	if m.ref == memory.ref {
		us := vm.Memory[s.Int() : s.Int()+n.Int()*4]
		r := *((*jsInt32Array)(unsafe.Pointer(&us)))
		return r
	}
	panic("jsInt32Array New invalid args")
}

type jsUint8Array []uint8

func (a jsUint8Array) New(vm *exec.VirtualMachine, args ...Value) interface{} {
	if len(args) != 3 {
		panic("jsUint8Array New len(args) != 3")
	}

	m := args[0]
	s := args[1]
	n := args[2]

	if s.IsNaNHead() || n.IsNaNHead() {
		return make(jsUint8Array, 0)
	}

	if m.ref == memory.ref {
		var r jsUint8Array = vm.Memory[s.Int() : s.Int()+n.Int()]
		return r
	}
	panic("jsUint8Array New invalid args")
}

type jsUint16Array []uint16

func (a jsUint16Array) New(vm *exec.VirtualMachine, args ...Value) interface{} {
	if len(args) != 3 {
		panic("jsUint16Array len(args) != 3")
	}

	m := args[0]
	s := args[1]
	n := args[2]

	if s.IsNaNHead() || n.IsNaNHead() {
		return make(jsUint16Array, 0)
	}

	if m.ref == memory.ref {
		us := vm.Memory[s.Int() : s.Int()+n.Int()*2]
		r := *((*jsUint16Array)(unsafe.Pointer(&us)))
		return r
	}
	panic("jsUint16Array New invalid args")
}

type jsUint32Array []uint32

func (a jsUint32Array) New(vm *exec.VirtualMachine, args ...Value) interface{} {
	if len(args) != 3 {
		panic("jsUint32Array New len(args) != 3")
	}

	m := args[0]
	s := args[1]
	n := args[2]

	if s.IsNaNHead() || n.IsNaNHead() {
		return make(jsUint32Array, 0)
	}

	if m.ref == memory.ref {
		us := vm.Memory[s.Int() : s.Int()+n.Int()*4]
		r := *((*jsUint32Array)(unsafe.Pointer(&us)))
		return r
	}
	panic("jsUint32Array New invalid args")
}

type jsFloat32Array []float32

func (a jsFloat32Array) New(vm *exec.VirtualMachine, args ...Value) interface{} {
	if len(args) != 3 {
		panic("jsFloat32Array New len(args) != 3")
	}

	m := args[0]
	s := args[1]
	n := args[2]

	if s.IsNaNHead() || n.IsNaNHead() {
		return make(jsFloat32Array, 0)
	}

	if m.ref == memory.ref {
		us := vm.Memory[s.Int() : s.Int()+n.Int()*4]
		r := *((*jsFloat32Array)(unsafe.Pointer(&us)))
		return r
	}
	panic("jsFloat32Array New invalid args")
}

type jsFloat64Array []float64

func (a jsFloat64Array) New(vm *exec.VirtualMachine, args ...Value) interface{} {
	if len(args) != 3 {
		panic("jsFloat64Array New len(args) != 3")
	}

	m := args[0]
	s := args[1]
	n := args[2]

	if s.IsNaNHead() || n.IsNaNHead() {
		return make(jsFloat64Array, 0)
	}

	if m.ref == memory.ref {
		us := vm.Memory[s.Int() : s.Int()+n.Int()*8]
		r := *((*jsFloat64Array)(unsafe.Pointer(&us)))
		return r
	}
	panic("jsFloat64Array New invalid args")
}

type JSArray []interface{}

func (a JSArray) Length() int {
	return len(a)
}

func (a JSArray) Get(index int64) interface{} {
	if len(a) <= int(index) {
		panic("JSArray: index out of range")
	}

	return a[index]
}

func (a JSArray) New(vm *exec.VirtualMachine, args ...Value) interface{} {
	if len(args) < 1 {
		return nil
	}
	result := make(JSArray, args[0].Int())
	return result
}

func (a JSArray) Set(index int64, x Value) {
	if len(a) <= int(index) {
		return
	}

	a[index] = x
}

func ByteSlice2JSArray(data []byte) JSArray {
	is := make(JSArray, len(data))
	for i, d := range data {
		is[i] = d
	}
	return is
}
