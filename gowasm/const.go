package gowasm

import (
	"fmt"
	"math"
	"os"
	"syscall"
)

const nanHead = 0x7FF80000

var (
	undefined = struct{}{}
	jsGlobal  = JSObject{
		"Array":        JSArray{},
		"Object":       JSObject{},
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

type stat struct {
	syscall.Stat_t
	JSObject
}

func newStat() *stat {
	s := &stat{}
	s.JSObject = JSObject{
		"isDirectory": func(args ...Value) interface{} {
			return s.Mode&syscall.S_IFMT == syscall.S_IFDIR
		},
	}
	return s
}

var fs = JSObject{
	"constants": JSObject{
		"O_WRONLY": float64(os.O_WRONLY),
		"O_RDWR":   float64(os.O_RDWR),
		"O_CREAT":  float64(os.O_CREATE),
		"O_TRUNC":  float64(os.O_TRUNC),
		"O_APPEND": float64(os.O_APPEND),
		"O_EXCL":   float64(os.O_EXCL),
	},
	"openSync": func(args ...Value) interface{} {
		if len(args) < 3 {
			return NewJSError(fmt.Errorf("fs.openSync: invalid num of args"))
		}
		path := args[0].String()
		mode := args[1].Int()
		perm := args[2].Uint32()
		n, err := syscall.Open(path, mode, perm)
		if err != nil {
			return NewJSError(err)
		}
		return n
	},
	"readSync": func(args ...Value) interface{} {
		if len(args) < 4 {
			return NewJSError(fmt.Errorf("fs.readSync: invalid num of args"))
		}
		fd := args[0].Int()
		buf := args[1].v.(jsUint8Array)
		s := args[2].Int()
		n := args[3].Int()
		n, err := syscall.Read(fd, buf[s:s+n])
		if err != nil {
			return NewJSError(err)
		}
		return n
	},
	"statSync": func(args ...Value) interface{} {
		if len(args) < 1 {
			return NewJSError(fmt.Errorf("fs.statSync: invalid num of args"))
		}
		path := args[0].String()
		stat := newStat()
		err := syscall.Stat(path, &stat.Stat_t)
		if err != nil {
			return NewJSError(err)
		}
		return stat
	},
	"fstatSync": func(args ...Value) interface{} {
		if len(args) < 1 {
			return NewJSError(fmt.Errorf("fs.fstatSync: invalid num of args"))
		}
		fd := args[0].Int()
		stat := newStat()
		err := syscall.Fstat(fd, &stat.Stat_t)
		if err != nil {
			return NewJSError(err)
		}
		return stat
	},
	"writeSync": func(args ...Value) interface{} {
		if len(args) < 4 {
			return NewJSError(fmt.Errorf("fs.writeSync: invalid num of args"))
		}
		fd := args[0].Int()
		buf := args[1].v.(jsUint8Array)
		s := args[2].Int()
		n := args[3].Int()
		n, err := syscall.Write(int(fd), buf[s:s+n])
		if err != nil {
			return NewJSError(err)
		}
		return n
	},
	"closeSync": func(args ...Value) interface{} {
		if len(args) < 1 {
			return NewJSError(fmt.Errorf("fs.closeSync: invalid num of args"))
		}
		fd := args[0].Int()
		err := syscall.Close(int(fd))
		if err != nil {
			return NewJSError(err)
		}
		return nil
	},
}
