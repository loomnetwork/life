package gowasm

import "github.com/perlin-network/life/exec"

type jsInt8Array []int8

func (a jsInt8Array) New(vm *exec.VirtualMachine,args ...interface{})  interface{} {
	r := make(jsInt8Array, len(args))
	for i, v := range args {
		r[i] = v.(int8)
	}
	return r
}

type jsInt16Array []int16

func (a jsInt16Array) New(vm *exec.VirtualMachine,args ...interface{})  interface{} {
	r := make(jsInt16Array, len(args))
	for i, v := range args {
		r[i] = v.(int16)
	}
	return r
}

type jsInt32Array []int32

func (a jsInt32Array) New(vm *exec.VirtualMachine,args ...interface{})  interface{} {
	r := make(jsInt32Array, len(args))
	for i, v := range args {
		r[i] = v.(int32)
	}
	return r
}

type jsUint8Array []uint8

func (a jsUint8Array) New(vm *exec.VirtualMachine,args ...interface{})  interface{} {
	if len(args) != 3 {
		panic("len(args) != 3")
	}

	m := args[0].(Value)
	s := args[1].(Value)
	n := args[2].(Value)

	if s.ref>>32&nanHead == nanHead || n.ref>>32&nanHead == nanHead {
		return make(jsUint8Array, 0)
	}

	if m.ref == memory.ref {
		len := int(n.v.(float64))
		r := make(jsUint8Array, len)
		start := int(s.v.(float64))
		copy(r, vm.Memory[start:])
		return r
	}
	panic("jsUint8Array New invalid args")
	return nil
}

type jsUint16Array []uint16

func (a jsUint16Array) New(vm *exec.VirtualMachine,args ...interface{})  interface{} {
	r := make(jsUint16Array, len(args))
	for i, v := range args {
		r[i] = v.(uint16)
	}
	return r
}

type jsUint32Array []uint32

func (a jsUint32Array) New(vm *exec.VirtualMachine,args ...interface{})  interface{} {
	r := make(jsUint32Array, len(args))
	for i, v := range args {
		r[i] = v.(uint32)
	}
	return r
}

type jsFloat32Array []float32

func (a jsFloat32Array) New(vm *exec.VirtualMachine,args ...interface{})  interface{} {
	r := make(jsFloat32Array, len(args))
	for i, v := range args {
		r[i] = v.(float32)
	}
	return r
}

type jsFloat64Array []float64

func (a jsFloat64Array) New(vm *exec.VirtualMachine,args ...interface{})  interface{} {
	r := make(jsFloat64Array, len(args))
	for i, v := range args {
		r[i] = v.(float64)
	}
	return r
}

type jsArray []interface{}

func (a jsArray) New(vm *exec.VirtualMachine,args ...interface{})  interface{} {
	return jsArray(args)
}

