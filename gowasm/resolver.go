package gowasm

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"github.com/perlin-network/life/exec"
	"io"
	"math"
	"os"
	"reflect"
	"syscall"
	"time"
)

type Resolver struct {
	StandAlone bool
	origin     time.Time

	values     map[ref]Value
	valueIndex ref
}

func NewResolver() *Resolver {
	return &Resolver{
		StandAlone: true,
		origin:     time.Now(),
		values:     cloneDefaultValues(),
		valueIndex: ref(len(defaultValues)),
	}
}

func cloneDefaultValues() (clone map[ref]Value) {
	clone = make(map[ref]Value, len(defaultValues))
	for key, value := range defaultValues {
		clone[key] = value
	}
	return
}

func (r *Resolver) Reset() {
	r.values = cloneDefaultValues()
	r.valueIndex = ref(len(defaultValues))
}

func (r *Resolver) Clone() exec.ImportResolver {
	return &Resolver{
		StandAlone: r.StandAlone,
		origin:     time.Now(),
		values:     cloneDefaultValues(),
		valueIndex: ref(len(defaultValues)),
	}
}

func (r *Resolver) SetGlobalValue(key string, value JSObject) {
	r.values[valueGlobal.RefLower32Bits()].v.(JSObject)[key] = value
}

func (r *Resolver) ResolveFunc(module, field string) exec.FunctionImport {
	switch module {
	case "go":
		switch field {
		case "runtime.getRandomData":
			return func(vm *exec.VirtualMachine) int64 {
				frame := vm.GetCurrentFrame()
				ptr := int(frame.Locals[0]) + 8
				base := binary.LittleEndian.Uint64(vm.Memory[ptr:])
				len := binary.LittleEndian.Uint64(vm.Memory[ptr+8:])
				_, err := rand.Read(vm.Memory[base : base+len])
				if err != nil {
					panic(fmt.Sprintf("runtime.getRandomData: err %v", err))
				}
				return 0
			}
		case "runtime.nanotime":
			return func(vm *exec.VirtualMachine) int64 {
				frame := vm.GetCurrentFrame()
				ptr := int(frame.Locals[0]) + 8
				nano := time.Since(r.origin).Nanoseconds()
				binary.LittleEndian.PutUint64(vm.Memory[ptr:], uint64(nano))
				return 0
			}
		case "runtime.walltime":
			return func(vm *exec.VirtualMachine) int64 {
				frame := vm.GetCurrentFrame()
				ptr := int(frame.Locals[0]) + 8
				now := time.Now()
				sec := now.Unix()
				binary.LittleEndian.PutUint64(vm.Memory[ptr:], uint64(sec))
				binary.LittleEndian.PutUint32(vm.Memory[ptr+8:], uint32(now.Nanosecond()))
				return 0
			}
		case "runtime.wasmExit":
			return func(vm *exec.VirtualMachine) int64 {
				if !r.StandAlone {
					return 0
				}
				frame := vm.GetCurrentFrame()
				ptr := int(frame.Locals[0]) + 8
				code := binary.LittleEndian.Uint32(vm.Memory[ptr:])
				os.Exit(int(code))
				return 0
			}
		case "runtime.wasmWrite":
			return func(vm *exec.VirtualMachine) int64 {
				frame := vm.GetCurrentFrame()
				ptr := int(frame.Locals[0]) + 8
				fd := binary.LittleEndian.Uint64(vm.Memory[ptr:])
				base := binary.LittleEndian.Uint64(vm.Memory[ptr+8:])
				len := binary.LittleEndian.Uint64(vm.Memory[ptr+16:])

				var writer io.Writer
				switch int(fd) {
				case syscall.Stdout:
					writer = os.Stdout
				case syscall.Stderr:
					writer = os.Stderr
				default:
					panic(fmt.Sprintf("runtime.wasmWrite: invalid fd %d", fd))
				}
				_, err := io.WriteString(writer, string(vm.Memory[base:base+len]))
				if err != nil {
					panic(fmt.Sprintf("runtime.wasmWrite: err %v", err))
				}

				return 0
			}
		case "syscall/js.valueGet":
			return func(vm *exec.VirtualMachine) int64 {
				frame := vm.GetCurrentFrame()
				ptr := int(frame.Locals[0]) + 8
				v := r.loadValue(vm, ptr)
				s := r.loadString(vm, ptr+8)
				gt, ok := v.v.(getter)
				if !ok {
					r.storeValue(vm, ptr+24, v)
				} else {
					r.storeValue(vm, ptr+24, gt.Get(s))
				}

				return 0
			}
		case "syscall/js.valueLength":
			return func(vm *exec.VirtualMachine) int64 {
				frame := vm.GetCurrentFrame()
				ptr := int(frame.Locals[0]) + 8

				v := r.loadValue(vm, ptr)
				switch value := v.v.(type) {
				case []interface{}:
					binary.LittleEndian.PutUint64(vm.Memory[ptr+8:], uint64(len(value)))
				case lengthGetter:
					binary.LittleEndian.PutUint64(vm.Memory[ptr+8:], uint64(value.Length()))
				default:
					fmt.Printf("syscall/js.valueLength error on %#v\n", v)
				}
				return 0
			}
		case "syscall/js.valueIndex":
			return func(vm *exec.VirtualMachine) int64 {
				frame := vm.GetCurrentFrame()
				ptr := int(frame.Locals[0]) + 8

				v := r.loadValue(vm, ptr)
				i := int64(binary.LittleEndian.Uint64(vm.Memory[ptr+8:]))
				switch value := v.v.(type) {
				case []interface{}:
					r.storeValue(vm, ptr+16, value[i])
				case indexGetter:
					r.storeValue(vm, ptr+16, value.Get(i))
				default:
					fmt.Printf("syscall/js.valueIndex error on %#v\n", v)
				}
				return 0
			}
		case "syscall/js.valueSetIndex":
			return func(vm *exec.VirtualMachine) int64 {
				frame := vm.GetCurrentFrame()
				ptr := int(frame.Locals[0]) + 8

				v := r.loadValue(vm, ptr)
				i := int64(binary.LittleEndian.Uint64(vm.Memory[ptr+8:]))
				x := r.loadValue(vm, ptr+16)
				switch value := v.v.(type) {
				case []interface{}:
					value[i] = x
				case indexSetter:
					value.Set(i, x)
				default:
					fmt.Printf("syscall/js.valueSetIndex error on %#v\n", v)
				}
				return 0
			}
		case "syscall/js.valueNew":
			return func(vm *exec.VirtualMachine) int64 {
				frame := vm.GetCurrentFrame()
				ptr := int(frame.Locals[0]) + 8
				v := r.loadValue(vm, ptr)
				values := r.loadSliceOfValues(vm, ptr+8)
				n, ok := v.v.(newer)
				if !ok {
					r.storeValue(vm, ptr+32, v)
					vm.Memory[ptr+40] = 1
				} else {
					r.storeValue(vm, ptr+32, n.New(vm, values...))
					vm.Memory[ptr+40] = 1
				}

				return 0
			}
		case "syscall/js.valuePrepareString":
			return func(vm *exec.VirtualMachine) int64 {
				frame := vm.GetCurrentFrame()
				ptr := int(frame.Locals[0]) + 8
				v := r.loadValue(vm, ptr)
				r.storeValue(vm, ptr+8, v)
				if s, ok := v.v.(string); ok {
					binary.LittleEndian.PutUint64(vm.Memory[ptr+16:], uint64(len(s)))
				} else {
					binary.LittleEndian.PutUint64(vm.Memory[ptr+16:], 0)
				}

				return 0
			}
		case "syscall/js.valueLoadString":
			return func(vm *exec.VirtualMachine) int64 {
				frame := vm.GetCurrentFrame()
				ptr := int(frame.Locals[0]) + 8
				v := r.loadValue(vm, ptr)
				if s, ok := v.v.(string); ok {
					copy(r.loadSlice(vm, ptr+8), []byte(s))
				} else {
					slice := r.loadSlice(vm, ptr+8)
					copy(slice, make([]byte, len(slice)))
				}

				return 0
			}
		case "syscall/js.valueCall":
			return func(vm *exec.VirtualMachine) int64 {
				frame := vm.GetCurrentFrame()
				ptr := int(frame.Locals[0]) + 8
				v := r.loadValue(vm, ptr)
				s := r.loadString(vm, ptr+8)
				args := r.loadSliceOfValues(vm, ptr+24)

				ptr = ptr + 48
				if c, ok := v.v.(caller); ok {
					ret := c.Call(s, args...)
					r.storeValue(vm, ptr, ret)
					if _, ok := ret.(jsError); ok {
						vm.Memory[ptr+8] = 0
					} else {
						vm.Memory[ptr+8] = 1
					}
				} else {
					r.storeValue(vm, ptr, NewJSError(fmt.Errorf("value %#v is not caller", v)))
					vm.Memory[ptr+8] = 0
				}
				return 0
			}
		case "syscall/js.stringVal":
			return func(vm *exec.VirtualMachine) int64 {
				frame := vm.GetCurrentFrame()
				ptr := int(frame.Locals[0]) + 8
				s := r.loadString(vm, ptr)
				r.storeValue(vm, ptr+16, s)
				return 0
			}
		default:
			return func(vm *exec.VirtualMachine) int64 {
				fmt.Printf("[app] called %s %s\n", module, field)
				return 0
			}
		}
	default:
		return func(vm *exec.VirtualMachine) int64 {
			fmt.Printf("[app] called %s %s\n", module, field)
			return 0
		}
	}
}

func (r *Resolver) loadSlice(vm *exec.VirtualMachine, ptr int) []byte {
	base := binary.LittleEndian.Uint64(vm.Memory[ptr:])
	len := binary.LittleEndian.Uint64(vm.Memory[ptr+8:])
	return vm.Memory[base : base+len]
}

func (r *Resolver) loadString(vm *exec.VirtualMachine, ptr int) string {
	return string(r.loadSlice(vm, ptr))
}

func (r *Resolver) loadValue(vm *exec.VirtualMachine, ptr int) Value {
	u := binary.LittleEndian.Uint64(vm.Memory[ptr:])
	if u>>32&nanHead != nanHead {
		return makeValue(ref(u), math.Float64frombits(u))
	}

	if u == valueNaN.RefUint64() {
		return valueNaN
	}

	k := ref(u).Lower32Bits()

	v, ok := r.values[k]
	if !ok {
		return valueNaN
	}
	return v
}

func (r *Resolver) loadSliceOfValues(vm *exec.VirtualMachine, ptr int) []Value {
	base := binary.LittleEndian.Uint64(vm.Memory[ptr:])
	len := binary.LittleEndian.Uint64(vm.Memory[ptr+8:])

	slice := make([]Value, len)
	for i := uint64(0); i < len; i++ {
		slice[i] = r.loadValue(vm, int(base+i*8))
	}
	return slice
}

const (
	typeFlagString   = 1
	typeFlagSymbol   = 2
	typeFlagFunction = 3
)

func (r *Resolver) storeValue(vm *exec.VirtualMachine, ptr int, v interface{}) {
	if v == nil {
		binary.LittleEndian.PutUint64(vm.Memory[ptr:], valueNull.RefUint64())
		return
	}

	switch tv := v.(type) {
	case byte:
		binary.LittleEndian.PutUint64(vm.Memory[ptr:], math.Float64bits(float64(tv)))
		return
	case int:
		binary.LittleEndian.PutUint64(vm.Memory[ptr:], math.Float64bits(float64(tv)))
		return
	case int64:
		binary.LittleEndian.PutUint64(vm.Memory[ptr:], math.Float64bits(float64(tv)))
		return
	case float64:
		if math.IsNaN(tv) {
			binary.LittleEndian.PutUint64(vm.Memory[ptr:], valueNaN.RefUint64())
			return
		}
		binary.LittleEndian.PutUint64(vm.Memory[ptr:], math.Float64bits(tv))
		return
	case bool:
		if tv {
			binary.LittleEndian.PutUint64(vm.Memory[ptr:], valueTrue.RefUint64())
		} else {
			binary.LittleEndian.PutUint64(vm.Memory[ptr:], valueFalse.RefUint64())
		}
		return
	case Value:
		binary.LittleEndian.PutUint64(vm.Memory[ptr:], tv.RefUint64())
		return
	default:
		var tf uint32
		switch reflect.TypeOf(v).Kind() {
		case reflect.String:
			tf = typeFlagString
		case reflect.Func:
			tf = typeFlagFunction
		}

		nr := ref(nanHead|tf)<<32 | r.valueIndex
		binary.LittleEndian.PutUint64(vm.Memory[ptr:], uint64(nr))

		r.values[r.valueIndex] = makeValue(nr, tv)
		r.valueIndex++
	}
}

func (r *Resolver) ResolveGlobal(module, field string) int64 {
	panic(fmt.Sprintf("Can not resolve global: %s %s\n", module, field))
}
