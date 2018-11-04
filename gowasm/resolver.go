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
	origin time.Time

	values     map[ref]Value
	valueIndex ref
}

func NewResolver() *Resolver {
	return &Resolver{
		origin:     time.Now(),
		values:     defaultValues,
		valueIndex: ref(len(defaultValues)),
	}
}

func (r *Resolver) ResolveFunc(module, field string) exec.FunctionImport {
	switch module {
	case "go":
		switch field {
		case "runtime.getRandomData":
			return func(vm *exec.VirtualMachine) int64 {
				frame := vm.GetCurrentFrame()
				ptr := int(uint32(frame.Locals[0])) + 8
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
				ptr := int(uint32(frame.Locals[0])) + 8
				nano := time.Since(r.origin).Nanoseconds()
				binary.LittleEndian.PutUint64(vm.Memory[ptr:], uint64(nano))
				return 0
			}
		case "runtime.walltime":
			return func(vm *exec.VirtualMachine) int64 {
				frame := vm.GetCurrentFrame()
				ptr := int(uint32(frame.Locals[0])) + 8
				now := time.Now()
				sec := now.Unix()
				binary.LittleEndian.PutUint64(vm.Memory[ptr:], uint64(sec))
				binary.LittleEndian.PutUint32(vm.Memory[ptr+8:], uint32(now.Nanosecond()))
				return 0
			}
		case "runtime.wasmExit":
			return func(vm *exec.VirtualMachine) int64 {
				frame := vm.GetCurrentFrame()
				ptr := int(uint32(frame.Locals[0])) + 8
				code := binary.LittleEndian.Uint32(vm.Memory[ptr:])
				os.Exit(int(code))
				return 0
			}
		case "runtime.wasmWrite":
			return func(vm *exec.VirtualMachine) int64 {
				frame := vm.GetCurrentFrame()
				ptr := int(uint32(frame.Locals[0])) + 8
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
				ptr := int(uint32(frame.Locals[0])) + 8
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
		case "syscall/js.valueNew":
			return func(vm *exec.VirtualMachine) int64 {
				frame := vm.GetCurrentFrame()
				ptr := int(uint32(frame.Locals[0])) + 8
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
				ptr := int(uint32(frame.Locals[0])) + 8
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
				ptr := int(uint32(frame.Locals[0])) + 8
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
				ptr := int(uint32(frame.Locals[0])) + 8
				v := r.loadValue(vm, ptr)
				s := r.loadString(vm, ptr+8)
				args := r.loadSliceOfValues(vm, ptr+24)

				if c, ok := v.v.(caller); ok {
					r.storeValue(vm, ptr+56, c.Call(s, args...))
					vm.Memory[ptr+64] = 1
				} else {
					r.storeValue(vm, ptr+56, newJSError(fmt.Sprintf("value %#v is not caller", v)))
					vm.Memory[ptr+64] = 0
				}
				return 0
			}
		case "syscall/js.stringVal":
			return func(vm *exec.VirtualMachine) int64 {
				frame := vm.GetCurrentFrame()
				ptr := int(uint32(frame.Locals[0])) + 8
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

func (r *Resolver) loadSliceOfValues(vm *exec.VirtualMachine, ptr int) []interface{} {
	base := binary.LittleEndian.Uint64(vm.Memory[ptr:])
	len := binary.LittleEndian.Uint64(vm.Memory[ptr+8:])

	slice := make([]interface{}, len)
	for i := uint64(0); i < len; i++ {
		slice[i] = r.loadValue(vm, int(base+i*8))
	}
	return slice
}

func (r *Resolver) storeValue(vm *exec.VirtualMachine, ptr int, v interface{}) {
	if v == nil {
		binary.LittleEndian.PutUint64(vm.Memory[ptr:], valueNull.RefUint64())
		return
	}

	var ref ref
	switch tv := v.(type) {
	case float64:
		if math.IsNaN(tv) {
			binary.LittleEndian.PutUint64(vm.Memory[ptr:], valueNaN.RefUint64())
			return
		}
		binary.LittleEndian.PutUint64(vm.Memory[ptr:], math.Float64bits(tv))
	case bool:
		if tv {
			binary.LittleEndian.PutUint64(vm.Memory[ptr:], valueTrue.RefUint64())
		} else {
			binary.LittleEndian.PutUint64(vm.Memory[ptr:], valueFalse.RefUint64())
		}
		return
	case Value:
		if tv.ref>>32&nanHead == nanHead {
			binary.LittleEndian.PutUint64(vm.Memory[ptr:], tv.RefUint64())
			return
		}
		value := reflect.ValueOf(tv.v)
		k := value.Kind()
		switch k {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			binary.LittleEndian.PutUint64(vm.Memory[ptr:], uint64(value.Int()))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			binary.LittleEndian.PutUint64(vm.Memory[ptr:], value.Uint())
		case reflect.Float32, reflect.Float64:
			binary.LittleEndian.PutUint64(vm.Memory[ptr:], math.Float64bits(value.Float()))
		default:
			binary.LittleEndian.PutUint64(vm.Memory[ptr:], tv.RefUint64())
		}
		return
	default:
		ref = r.valueIndex
		r.values[ref] = makeValue(ref, tv)
		r.valueIndex++
	}

	const (
		typeFlagString   = 1
		typeFlagSymbol   = 2
		typeFlagFunction = 3
	)

	var tf uint32
	switch reflect.TypeOf(v).Kind() {
	case reflect.String:
		tf = typeFlagString
	case reflect.Func:
		tf = typeFlagFunction
	}

	binary.LittleEndian.PutUint32(vm.Memory[ptr+4:], nanHead|tf)
	binary.LittleEndian.PutUint32(vm.Memory[ptr:], uint32(ref))
}

func (r *Resolver) ResolveGlobal(module, field string) int64 {
	panic(fmt.Sprintf("Can not resolve global: %s %s\n", module, field))
}
