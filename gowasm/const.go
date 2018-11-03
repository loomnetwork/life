package gowasm

import (
	"fmt"
	"math"
	"os"
)

const nanHead = 0x7FF80000

var (
	undefined = struct{}{}
	jsGlobal  = jsObject{
		"Array":        jsArray{},
		"Object":       jsObject{},
		"Int8Array":    jsInt8Array{},
		"Int16Array":   jsInt16Array{},
		"Int32Array":   jsInt32Array{},
		"Uint8Array":   jsUint8Array{},
		"Uint16Array":  jsUint16Array{},
		"Uint32Array":  jsUint32Array{},
		"Float32Array": jsFloat32Array{},
		"Float64Array": jsFloat64Array{},
		"fs":           fs,
	}
)

var (
	valueNaN       = predefValue(0, math.NaN())
	valueUndefined = predefValue(1, undefined)
	valueNull      = predefValue(2, nil)
	valueTrue      = predefValue(3, true)
	valueFalse     = predefValue(4, false)
	valueGlobal    = predefValue(5, jsGlobal)
	memory         = predefValue(6, nil) // WebAssembly linear memory
	jsGo           = predefValue(7, nil) // instance of the Go class in JavaScript
)

var defaultValues = map[ref]Value{
	valueNaN.RefLower32Bits():       valueNaN,
	valueUndefined.RefLower32Bits(): valueUndefined,
	valueNull.RefLower32Bits():      valueNull,
	valueTrue.RefLower32Bits():      valueTrue,
	valueFalse.RefLower32Bits():     valueFalse,
	valueGlobal.RefLower32Bits():    valueGlobal,
	memory.RefLower32Bits():         memory,
	jsGo.RefLower32Bits():           jsGo,
}

var fs = jsObject{
	"constants": jsObject{
		"O_WRONLY": float64(os.O_WRONLY),
		"O_RDWR":   float64(os.O_RDWR),
		"O_CREAT":  float64(os.O_CREATE),
		"O_TRUNC":  float64(os.O_TRUNC),
		"O_APPEND": float64(os.O_APPEND),
		"O_EXCL":   float64(os.O_EXCL),
	},
	"writeSync": func(args ...interface{}) interface{} {
		if len(args) < 4 {
			return nil
		}
		bufValue := args[1].(Value)
		buf := bufValue.v.(jsUint8Array)
		// s := args[2]
		// n := args[3]
		fmt.Print(string(buf))
		return nil
	},
}
