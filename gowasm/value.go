package gowasm

import (
	"fmt"
	"github.com/perlin-network/life/exec"
)

type getter interface {
	Get(v string) interface{}
}

type newer interface {
	New(vm *exec.VirtualMachine, args ...Value) interface{}
}

type caller interface {
	Call(method string, args ...Value) interface{}
}

type ref uint64

func (r ref) Lower32Bits() ref {
	return r & 0xffffffff
}

func (r ref) String() string {
	return fmt.Sprintf("%x", uint64(r))
}

type Value struct {
	ref ref
	v   interface{}
}

func makeValue(r ref, v interface{}) Value {
	return Value{
		ref: r,
		v:   v,
	}
}

func predefValue(id uint32, value interface{}) Value {
	return Value{ref: nanHead<<32 | ref(id), v: value}
}

func (v Value) RefUint64() uint64 {
	return uint64(v.ref)
}

func (v Value) RefLower32Bits() ref {
	return v.ref & 0xffffffff
}

func (v Value) IsNaNHead() bool {
	return v.ref>>32&nanHead == nanHead
}

func (v Value) String() string {
	return v.v.(string)
}

func (v Value) float() float64 {
	if !v.isNumber() {
		panic("Value is not number type")
	}
	return v.v.(float64)
}

func (v Value) Float() float64 {
	return v.float()
}

func (v Value) Int() int {
	return int(v.float())
}

func (v Value) Uint32() uint32 {
	return uint32(v.float())
}

type Type int

const (
	TypeUndefined Type = iota
	TypeNull
	TypeBoolean
	TypeNumber
	TypeString
	TypeSymbol
	TypeObject
	TypeFunction
)

func (t Type) String() string {
	switch t {
	case TypeUndefined:
		return "undefined"
	case TypeNull:
		return "null"
	case TypeBoolean:
		return "boolean"
	case TypeNumber:
		return "number"
	case TypeString:
		return "string"
	case TypeSymbol:
		return "symbol"
	case TypeObject:
		return "object"
	case TypeFunction:
		return "function"
	default:
		panic("bad type")
	}
}

func (v Value) isNumber() bool {
	return v.ref>>32&nanHead != nanHead || v.ref == valueNaN.ref
}

func (v Value) Type() Type {
	switch v.ref {
	case valueUndefined.ref:
		return TypeUndefined
	case valueNull.ref:
		return TypeNull
	case valueTrue.ref, valueFalse.ref:
		return TypeBoolean
	}
	if v.isNumber() {
		return TypeNumber
	}
	typeFlag := v.ref >> 32 & 3
	switch typeFlag {
	case 1:
		return TypeString
	case 2:
		return TypeSymbol
	case 3:
		return TypeFunction
	default:
		return TypeObject
	}
}
