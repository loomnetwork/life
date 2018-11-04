package gowasm

import (
	"encoding/binary"
	"github.com/perlin-network/life/exec"
	"math"
)

type jsInt8Array []int8

func (a jsInt8Array) New(vm *exec.VirtualMachine, args ...interface{}) interface{} {
	if len(args) != 3 {
		panic("jsInt8Array New len(args) != 3")
	}

	m := args[0].(Value)
	s := args[1].(Value)
	n := args[2].(Value)

	if s.IsNaNHead() || n.IsNaNHead() {
		return make(jsInt8Array, 0)
	}

	if m.ref == memory.ref {
		r := make(jsInt8Array, n.Int())
		s := s.Int()
		for i := range r {
			r[i] = int8(vm.Memory[s+i])
		}
		return r
	}
	panic("jsInt8Array New invalid args")
}

type jsInt16Array []int16

func (a jsInt16Array) New(vm *exec.VirtualMachine, args ...interface{}) interface{} {
	if len(args) != 3 {
		panic("jsInt16Array New len(args) != 3")
	}

	m := args[0].(Value)
	s := args[1].(Value)
	n := args[2].(Value)

	if s.IsNaNHead() || n.IsNaNHead() {
		return make(jsInt16Array, 0)
	}

	if m.ref == memory.ref {
		r := make(jsInt16Array, n.Int())
		s := s.Int()
		for i := range r {
			r[i] = int16(binary.LittleEndian.Uint16(vm.Memory[s+i*2:]))
		}
		return r
	}
	panic("jsInt16Array New invalid args")
}

type jsInt32Array []int32

func (a jsInt32Array) New(vm *exec.VirtualMachine, args ...interface{}) interface{} {
	if len(args) != 3 {
		panic("jsInt32Array New len(args) != 3")
	}

	m := args[0].(Value)
	s := args[1].(Value)
	n := args[2].(Value)

	if s.IsNaNHead() || n.IsNaNHead() {
		return make(jsInt32Array, 0)
	}

	if m.ref == memory.ref {
		r := make(jsInt32Array, n.Int())
		s := s.Int()
		for i := range r {
			r[i] = int32(binary.LittleEndian.Uint16(vm.Memory[s+i*4:]))
		}
		return r
	}
	panic("jsInt32Array New invalid args")
}

type jsUint8Array []uint8

func (a jsUint8Array) New(vm *exec.VirtualMachine, args ...interface{}) interface{} {
	if len(args) != 3 {
		panic("jsUint8Array New len(args) != 3")
	}

	m := args[0].(Value)
	s := args[1].(Value)
	n := args[2].(Value)

	if s.IsNaNHead() || n.IsNaNHead() {
		return make(jsUint8Array, 0)
	}

	if m.ref == memory.ref {
		r := make(jsUint8Array, n.Int())
		copy(r, vm.Memory[s.Int():])
		return r
	}
	panic("jsUint8Array New invalid args")
}

type jsUint16Array []uint16

func (a jsUint16Array) New(vm *exec.VirtualMachine, args ...interface{}) interface{} {
	if len(args) != 3 {
		panic("jsUint16Array len(args) != 3")
	}

	m := args[0].(Value)
	s := args[1].(Value)
	n := args[2].(Value)

	if s.IsNaNHead() || n.IsNaNHead() {
		return make(jsUint16Array, 0)
	}

	if m.ref == memory.ref {
		r := make(jsUint16Array, n.Int())
		s := s.Int()
		for i := range r {
			r[i] = binary.LittleEndian.Uint16(vm.Memory[s+i*2:])
		}
		return r
	}
	panic("jsUint16Array New invalid args")
}

type jsUint32Array []uint32

func (a jsUint32Array) New(vm *exec.VirtualMachine, args ...interface{}) interface{} {
	if len(args) != 3 {
		panic("jsUint32Array New len(args) != 3")
	}

	m := args[0].(Value)
	s := args[1].(Value)
	n := args[2].(Value)

	if s.IsNaNHead() || n.IsNaNHead() {
		return make(jsUint32Array, 0)
	}

	if m.ref == memory.ref {
		r := make(jsUint32Array, n.Int())
		s := s.Int()
		for i := range r {
			r[i] = binary.LittleEndian.Uint32(vm.Memory[s+i*4:])
		}
		return r
	}
	panic("jsUint32Array New invalid args")
}

type jsFloat32Array []float32

func (a jsFloat32Array) New(vm *exec.VirtualMachine, args ...interface{}) interface{} {
	if len(args) != 3 {
		panic("jsFloat32Array New len(args) != 3")
	}

	m := args[0].(Value)
	s := args[1].(Value)
	n := args[2].(Value)

	if s.IsNaNHead() || n.IsNaNHead() {
		return make(jsFloat32Array, 0)
	}

	if m.ref == memory.ref {
		r := make(jsFloat32Array, n.Int())
		s := s.Int()
		for i := range r {
			r[i] = math.Float32frombits(binary.LittleEndian.Uint32(vm.Memory[s+i*4:]))
		}
		return r
	}
	panic("jsFloat32Array New invalid args")
}

type jsFloat64Array []float64

func (a jsFloat64Array) New(vm *exec.VirtualMachine, args ...interface{}) interface{} {
	if len(args) != 3 {
		panic("jsFloat64Array New len(args) != 3")
	}

	m := args[0].(Value)
	s := args[1].(Value)
	n := args[2].(Value)

	if s.IsNaNHead() || n.IsNaNHead() {
		return make(jsFloat64Array, 0)
	}

	if m.ref == memory.ref {
		r := make(jsFloat64Array, n.Int())
		s := s.Int()
		for i := range r {
			r[i] = math.Float64frombits(binary.LittleEndian.Uint64(vm.Memory[s+i*8:]))
		}
		return r
	}
	panic("jsFloat64Array New invalid args")
}

type jsArray []interface{}

func (a jsArray) New(vm *exec.VirtualMachine, args ...interface{}) interface{} {
	return jsArray(args)
}
