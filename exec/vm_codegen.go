package exec

import (
	"fmt"
	"math"
	"github.com/perlin-network/life/compiler/opcodes"
	"github.com/perlin-network/life/compiler"
)

type jitContext struct {
	vm *VirtualMachine
	functionID int
	code *compiler.InterpreterCode
	program string
	cont int
	ip int
	thisIP int
}

func (c *jitContext) writeFallback() {
	c.program += fmt.Sprintf("*ret = %d;\n", c.thisIP)
	c.program += fmt.Sprintf("return %d;\n", c.cont)
	c.program += fmt.Sprintf("case %d:\n", c.cont)
	c.cont++
}

func (c *jitContext) checkLocal(id int) {
	if id < 0 || id >= c.code.NumParams + c.code.NumLocals {
		panic(fmt.Errorf("local out of bounds: id = %d, n = %d", id, c.code.NumParams + c.code.NumLocals))
	}
}

func (c *jitContext) checkReg(id int) {
	if id < 0 || id >= c.code.NumRegs {
		panic(fmt.Errorf("reg out of bounds: id = %d, n = %d", id, c.code.NumRegs))
	}
}

func (c *jitContext) checkGlobal(id int) {
	if id < 0 || id >= len(c.vm.Globals) {
		panic(fmt.Errorf("global out of bounds: id = %d, n = %d", id, len(c.vm.Globals)))
	}
}

func (c *jitContext) writeSI32Op(valueID int, op string) {
	a := int(LE.Uint32(c.code.Bytes[c.ip : c.ip + 4]))
	c.checkReg(a)
	b := int(LE.Uint32(c.code.Bytes[c.ip + 4 : c.ip + 8]))
	c.checkReg(b)

	c.ip += 8

	if op == "/" || op == "%" {
		c.program += fmt.Sprintf("if((u32) regs[%d] == 0) return -5;\n", b)
		c.program += fmt.Sprintf("if((i32) regs[%d] == %d && (i32) regs[%d] == -1)", a, math.MinInt32, b)
		if op == "/" {
			c.program += "return -5;"
		} else {
			c.program += fmt.Sprintf("regs[%d] = 0; else ", valueID)
		}
	}

	c.program += fmt.Sprintf("regs[%d] = (i64)((i32) regs[%d] %s (i32) regs[%d]);\n", valueID, a, op, b)
}

func (c *jitContext) writeUI32Op(valueID int, op string) {
	a := int(LE.Uint32(c.code.Bytes[c.ip : c.ip + 4]))
	c.checkReg(a)
	b := int(LE.Uint32(c.code.Bytes[c.ip + 4 : c.ip + 8]))
	c.checkReg(b)

	c.ip += 8

	if op == "/" || op == "%" {
		c.program += fmt.Sprintf("if((u32) regs[%d] == 0) return -5;\n", b)
	}

	c.program += fmt.Sprintf("regs[%d] = (i64)((u32) regs[%d] %s (u32) regs[%d]);\n", valueID, a, op, b)
}

func (c *jitContext) writeSI64Op(valueID int, op string) {
	a := int(LE.Uint32(c.code.Bytes[c.ip : c.ip + 4]))
	c.checkReg(a)
	b := int(LE.Uint32(c.code.Bytes[c.ip + 4 : c.ip + 8]))
	c.checkReg(b)

	c.ip += 8

	if op == "/" || op == "%" {
		c.program += fmt.Sprintf("if(regs[%d] == 0) return -5;\n", b)
		c.program += fmt.Sprintf("if(regs[%d] == %d && regs[%d] == -1)", a, math.MinInt64, b)
		if op == "/" {
			c.program += "return -5;"
		} else {
			c.program += fmt.Sprintf("regs[%d] = 0; else ", valueID)
		}
	}

	c.program += fmt.Sprintf("regs[%d] = regs[%d] %s regs[%d];\n", valueID, a, op, b)
}

func (c *jitContext) writeUI64Op(valueID int, op string) {
	a := int(LE.Uint32(c.code.Bytes[c.ip : c.ip + 4]))
	c.checkReg(a)
	b := int(LE.Uint32(c.code.Bytes[c.ip + 4 : c.ip + 8]))
	c.checkReg(b)

	c.ip += 8

	if op == "/" || op == "%" {
		c.program += fmt.Sprintf("if((u32) regs[%d] == 0) return -5;\n", b)
	}

	c.program += fmt.Sprintf("regs[%d] = (u64) regs[%d] %s (u64) regs[%d];\n", valueID, a, op, b)
}

func (c *jitContext) writeMemoryLoad(valueID int, ty string) {
	offset := LE.Uint32(c.code.Bytes[c.ip+4 : c.ip+8])

	base := int(LE.Uint32(c.code.Bytes[c.ip+8:c.ip+12]))
	c.checkReg(base)

	c.ip += 12

	c.program += fmt.Sprintf("tempPtr0 = %dUL + (unsigned long) (u32) regs[%d];\n", offset, base)
	c.program += fmt.Sprintf("if(tempPtr0 >= (unsigned long) memory_len) return -4;\n")
	c.program += fmt.Sprintf("regs[%d] = (i64) *(%s*)((unsigned long) memory + tempPtr0);\n", valueID, ty)
}

func (c *jitContext) writeMemoryStore(ty string) {
	offset := LE.Uint32(c.code.Bytes[c.ip+4 : c.ip+8])

	base := int(LE.Uint32(c.code.Bytes[c.ip+8:c.ip+12]))
	c.checkReg(base)

	value := int(LE.Uint32(c.code.Bytes[c.ip+12:c.ip+16]))
	c.checkReg(value)

	c.ip += 16

	c.program += fmt.Sprintf("tempPtr0 = %dUL + (unsigned long) (u32) regs[%d];\n", offset, base)
	c.program += fmt.Sprintf("if(tempPtr0 >= (unsigned long) memory_len) return -4;\n")
	c.program += fmt.Sprintf("*(%s*)((unsigned long) memory + tempPtr0) = (%s) regs[%d];\n", ty, ty, value)
}

func (c *jitContext) Generate() bool {
	c.program = `
typedef char i8;
typedef short i16;
typedef int i32;
typedef long long i64;
typedef unsigned char u8;
typedef unsigned short u16;
typedef unsigned int u32;
typedef unsigned long long u64;
`

	// Returns -1 for done. The return value should have already be written in ret.
	// Return >= 0 for continuation. In this case, the instruction location should be
	// written in `ret` and only the current instruction will get interpreted.
	c.program += `
i32 run(
	i64 *regs,
	i64 *locals,
	i64 *globals,
	u8 *memory,
	i64 memory_len,
	i64 *yielded,
	i32 continuation,
	i64 *ret
) {
unsigned long tempPtr0;

switch(continuation) {
case 0:
`

	c.cont = 1
	c.ip = 0

	for {
		if c.ip == len(c.code.Bytes) {
			break
		}
		c.program += fmt.Sprintf("I%d:\n", c.ip)
		c.thisIP = c.ip

		valueID := int(LE.Uint32(c.code.Bytes[c.ip : c.ip+4]))
		c.checkReg(valueID)

		ins := opcodes.Opcode(c.code.Bytes[c.ip+4])
		c.ip += 5

		switch ins {
		case opcodes.Nop:
		case opcodes.Unreachable:
			c.program += "return -2;\n"

		case opcodes.Select:
			a := int(LE.Uint32(c.code.Bytes[c.ip : c.ip + 4]))
			c.checkReg(a)
			b := int(LE.Uint32(c.code.Bytes[c.ip + 4 : c.ip + 8]))
			c.checkReg(b)
			condReg := int(LE.Uint32(c.code.Bytes[c.ip + 8 : c.ip + 12]))
			c.checkReg(condReg)

			c.ip += 12

			c.program += fmt.Sprintf("regs[%d] = regs[%d] ? regs[%d] : regs[%d];", valueID, condReg, a, b)

		case opcodes.I32Const:
			imm := int64(LE.Uint32(c.code.Bytes[c.ip:c.ip+4]))
			c.ip += 4
			c.program += fmt.Sprintf("regs[%d] = %dLL;\n", valueID, imm)
		case opcodes.I32Add:
			c.writeUI32Op(valueID, "+")
		case opcodes.I32Sub:
			c.writeUI32Op(valueID, "-")
		case opcodes.I32Mul:
			c.writeUI32Op(valueID, "*")
		case opcodes.I32DivU:
			c.writeUI32Op(valueID, "/")
		case opcodes.I32DivS:
			c.writeSI32Op(valueID, "/")
		case opcodes.I32RemU:
			c.writeUI32Op(valueID, "%")
		case opcodes.I32RemS:
			c.writeSI32Op(valueID, "%")
		case opcodes.I32And:
			c.writeUI32Op(valueID, "&")
		case opcodes.I32Or:
			c.writeUI32Op(valueID, "|")
		case opcodes.I32Xor:
			c.writeUI32Op(valueID, "^")
		case opcodes.I32Eq:
			c.writeUI32Op(valueID, "==")
		case opcodes.I32Ne:
			c.writeUI32Op(valueID, "!=")
		case opcodes.I32LtU:
			c.writeUI32Op(valueID, "<")
		case opcodes.I32LtS:
			c.writeSI32Op(valueID, "<")
		case opcodes.I32LeU:
			c.writeUI32Op(valueID, "<=")
		case opcodes.I32LeS:
			c.writeSI32Op(valueID, "<=")
		case opcodes.I32GtU:
			c.writeUI32Op(valueID, ">")
		case opcodes.I32GtS:
			c.writeSI32Op(valueID, ">")
		case opcodes.I32GeU:
			c.writeUI32Op(valueID, ">=")
		case opcodes.I32GeS:
			c.writeSI32Op(valueID, ">=")
		case opcodes.I32EqZ:
			a := int(LE.Uint32(c.code.Bytes[c.ip : c.ip + 4]))
			c.checkReg(a)

			c.ip += 4

			c.program += fmt.Sprintf("regs[%d] = (i64) ((u32) regs[%d] == 0);\n", valueID, a)
		case opcodes.I32Shl:
			a := int(LE.Uint32(c.code.Bytes[c.ip : c.ip + 4]))
			c.checkReg(a)
			b := int(LE.Uint32(c.code.Bytes[c.ip + 4 : c.ip + 8]))
			c.checkReg(b)

			c.ip += 8
			c.program += fmt.Sprintf("regs[%d] = (i64)(((u32) regs[%d]) << (((u32) regs[%d]) %% 32));\n", valueID, a, b)
		case opcodes.I32ShrU:
			a := int(LE.Uint32(c.code.Bytes[c.ip : c.ip + 4]))
			c.checkReg(a)
			b := int(LE.Uint32(c.code.Bytes[c.ip + 4 : c.ip + 8]))
			c.checkReg(b)

			c.ip += 8
			c.program += fmt.Sprintf("regs[%d] = (i64)(((u32) regs[%d]) >> (((u32) regs[%d]) %% 32));\n", valueID, a, b)
		case opcodes.I32ShrS:
			a := int(LE.Uint32(c.code.Bytes[c.ip : c.ip + 4]))
			c.checkReg(a)
			b := int(LE.Uint32(c.code.Bytes[c.ip + 4 : c.ip + 8]))
			c.checkReg(b)

			c.ip += 8
			c.program += fmt.Sprintf("regs[%d] = (i64)(((i32) regs[%d]) >> (((i32) regs[%d]) %% 32));\n", valueID, a, b)
		case opcodes.I64Const:
			imm := int64(LE.Uint64(c.code.Bytes[c.ip:c.ip+8]))
			c.ip += 8
			c.program += fmt.Sprintf("regs[%d] = %dLL;\n", valueID, imm)
		case opcodes.I64Add:
			c.writeUI64Op(valueID, "+")
		case opcodes.I64Eq:
			c.writeUI64Op(valueID, "==")
		case opcodes.GetLocal:
			id := int(LE.Uint32(c.code.Bytes[c.ip : c.ip+4]))
			c.checkLocal(id)

			c.ip += 4
			c.program += fmt.Sprintf("regs[%d] = locals[%d];\n", valueID, id)
		case opcodes.SetLocal:
			id := int(LE.Uint32(c.code.Bytes[c.ip:c.ip+4]))
			c.checkLocal(id)

			val := int(LE.Uint32(c.code.Bytes[c.ip+4:c.ip+8]))
			c.checkReg(val)

			c.ip += 8
			c.program += fmt.Sprintf("locals[%d] = regs[%d];\n", id, val)
		case opcodes.GetGlobal:
			id := int(LE.Uint32(c.code.Bytes[c.ip : c.ip+4]))
			c.checkGlobal(id)

			c.ip += 4
			c.program += fmt.Sprintf("regs[%d] = globals[%d];\n", valueID, id)
		case opcodes.SetGlobal:
			id := int(LE.Uint32(c.code.Bytes[c.ip:c.ip+4]))
			c.checkGlobal(id)

			val := int(LE.Uint32(c.code.Bytes[c.ip+4:c.ip+8]))
			c.checkReg(val)

			c.ip += 8
			c.program += fmt.Sprintf("globals[%d] = regs[%d];\n", id, val)
		case opcodes.I32Load:
			c.writeMemoryLoad(valueID, "u32")
		case opcodes.I64Load:
			c.writeMemoryLoad(valueID, "u64")
		case opcodes.I32Load8U, opcodes.I64Load8U:
			c.writeMemoryLoad(valueID, "u8")
		case opcodes.I32Load8S, opcodes.I64Load8S:
			c.writeMemoryLoad(valueID, "i8")
		case opcodes.I32Load16U, opcodes.I64Load16U:
			c.writeMemoryLoad(valueID, "u16")
		case opcodes.I32Load16S, opcodes.I64Load16S:
			c.writeMemoryLoad(valueID, "i16")
		case opcodes.I64Load32U:
			c.writeMemoryLoad(valueID, "u32")
		case opcodes.I64Load32S:
			c.writeMemoryLoad(valueID, "i32")
		case opcodes.I32Store:
			c.writeMemoryStore("u32")
		case opcodes.I64Store:
			c.writeMemoryStore("u64")
		case opcodes.I32Store8, opcodes.I64Store8:
			c.writeMemoryStore("u8")
		case opcodes.I32Store16, opcodes.I64Store16:
			c.writeMemoryStore("u16")
		case opcodes.I64Store32:
			c.writeMemoryStore("u32")
		case opcodes.ReturnVoid:
			c.writeFallback()
		case opcodes.ReturnValue:
			c.ip += 4
			c.writeFallback()
		case opcodes.Call, opcodes.CallIndirect:
			c.ip += 4
			argCount := int(LE.Uint32(c.code.Bytes[c.ip:c.ip+4]))
			c.ip += 4
			c.ip += 4 * argCount
			c.writeFallback()
		case opcodes.Jmp:
			target := int(LE.Uint32(c.code.Bytes[c.ip : c.ip+4]))
			yieldReg := int(LE.Uint32(c.code.Bytes[c.ip+4 : c.ip+8]))
			c.checkReg(yieldReg)
			c.ip += 8

			c.program += fmt.Sprintf("*yielded = regs[%d];\n", yieldReg)
			c.program += fmt.Sprintf("goto I%d;\n", target)
		case opcodes.JmpIf:
			target := int(LE.Uint32(c.code.Bytes[c.ip : c.ip+4]))

			cond := int(LE.Uint32(c.code.Bytes[c.ip+4 : c.ip+8]))
			c.checkReg(cond)

			yieldReg := int(LE.Uint32(c.code.Bytes[c.ip+8 : c.ip+12]))
			c.checkReg(yieldReg)

			c.ip += 12

			c.program += fmt.Sprintf("if(regs[%d]) {\n", cond)
			c.program += fmt.Sprintf("*yielded = regs[%d];\n", yieldReg)
			c.program += fmt.Sprintf("goto I%d;\n", target)
			c.program += "}\n"
		case opcodes.JmpTable:
			targetCount := int(LE.Uint32(c.code.Bytes[c.ip : c.ip+4]))
			c.ip += 4

			targetsRaw := c.code.Bytes[c.ip : c.ip+4*targetCount]
			c.ip += 4 * targetCount

			defaultTarget := int(LE.Uint32(c.code.Bytes[c.ip : c.ip+4]))
			c.ip += 4

			condReg := int(LE.Uint32(c.code.Bytes[c.ip : c.ip+4]))
			c.checkReg(condReg)
			c.ip += 4

			yieldReg := int(LE.Uint32(c.code.Bytes[c.ip : c.ip+4]))
			c.checkReg(yieldReg)
			c.ip += 4

			c.program += fmt.Sprintf("*yielded = regs[%d];\n", yieldReg)
			c.program += fmt.Sprintf("switch(regs[%d]) {\n", condReg)
			for i := 0; i < targetCount; i++ {
				target := int(LE.Uint32(targetsRaw[i * 4 : i * 4 + 4]))
				c.program += fmt.Sprintf("case %d: goto I%d;\n", i, target)
			}
			c.program += fmt.Sprintf("default: goto I%d;\n", defaultTarget)
			c.program += "}"
		case opcodes.Phi:
			fmt.Sprintf("regs[%d] = *yielded\n", valueID)
		default:
			fmt.Printf("unsupported op: %s\n", ins.String())
			return false
		}
	}

	c.program += `
	break;
	}
	return -3; // invalid
}
	`
	//fmt.Println(c.program)
	c.code.JITInfo = CompileDynamicModule(c.program)
	return true
}

// Generate C code for the given function.
// Returns true if codegen succeeds, or false if the current function cannot be code-generated.
func (vm *VirtualMachine) GenerateCodeForFunction(functionID int) bool {
	code := &vm.FunctionCode[functionID]
	c := &jitContext {
		vm: vm,
		functionID: functionID,
		code: code,
	}
	return c.Generate()
}
