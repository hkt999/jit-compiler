package opcodes

import (
	"fmt"
	"sort"

	"github.com/bspaans/jit-compiler/asm/x86_64/encoding"
	. "github.com/bspaans/jit-compiler/asm/x86_64/encoding"
	"github.com/bspaans/jit-compiler/lib"
)

type OpcodeMap map[lib.Type]map[lib.Size][]*Opcode

func (o OpcodeMap) add(ty lib.Type, si lib.Size, op *Opcode) {
	arr, found := o[ty][si]
	if !found {
		arr = []*Opcode{}
	}
	arr = append(arr, op)
	o[ty][si] = arr
}

type OpcodeMaps []OpcodeMap

func (o OpcodeMaps) ResolveOpcode(operands []lib.Operand) *Opcode {
	picks := map[*Opcode]bool{}

	for i, opcodeMap := range o {
		oper := operands[i]
		if oper == nil {
			return nil
		}
		reg, isRegister := oper.(*Register)
		matches := opcodeMap[oper.Type()][oper.Width()]
		if len(matches) == 0 {
			return nil
		}
		newPick := map[*Opcode]bool{}
		for _, opcode := range matches {
			if (oper == encoding.Ah || oper == encoding.Ch || oper == encoding.Dh || oper == encoding.Bh) && (opcode.HasExtension(Rex) || opcode.HasExtension(RexW)) {
				continue
			}
			if (oper == encoding.Spl || oper == encoding.Bpl || oper == encoding.Sil || oper == encoding.Dil ||
				(isRegister && reg.Register >= 8)) && !(opcode.HasExtension(Rex) || opcode.HasExtension(RexW) || opcode.HasExtension(VEX128) || opcode.HasExtension(VEX256)) {
				continue
			}

			if i == 0 {
				newPick[opcode] = true
			} else {
				if picks[opcode] {
					newPick[opcode] = true
				}
			}
		}
		picks = newPick
	}
	opcodes := []*Opcode{}
	for pick, _ := range picks {
		opcodes = append(opcodes, pick)
	}

	sort.Slice(opcodes, func(i, j int) bool {
		return opcodes[i].Operands[0].Type < opcodes[j].Operands[0].Type

	})
	for _, _ = range picks {
		return opcodes[0]
	}
	return nil
}

func NewOpcodeMap() OpcodeMap {
	return map[lib.Type]map[lib.Size][]*Opcode{
		lib.T_Register:          map[lib.Size][]*Opcode{},
		lib.T_IndirectRegister:  map[lib.Size][]*Opcode{},
		lib.T_SIBRegister:       map[lib.Size][]*Opcode{},
		lib.T_DisplacedRegister: map[lib.Size][]*Opcode{},
		lib.T_RIPRelative:       map[lib.Size][]*Opcode{},
		lib.T_Uint8:             map[lib.Size][]*Opcode{},
		lib.T_Uint16:            map[lib.Size][]*Opcode{},
		lib.T_Uint32:            map[lib.Size][]*Opcode{},
		lib.T_Uint64:            map[lib.Size][]*Opcode{},
		lib.T_Int32:             map[lib.Size][]*Opcode{},
		lib.T_Float32:           map[lib.Size][]*Opcode{},
		lib.T_Float64:           map[lib.Size][]*Opcode{},
	}
}

func OpcodesToOpcodeMaps(opcodes []*Opcode, argCount int) OpcodeMaps {
	maps := make([]OpcodeMap, argCount)
	for i := 0; i < argCount; i++ {
		opcodeMap := OpcodesToOpcodeMap(opcodes, i)
		maps[i] = opcodeMap
	}
	return maps
}

func OpcodesToOpcodeMap(opcodes []*Opcode, operand int) OpcodeMap {
	opcodeMap := NewOpcodeMap()
	for _, opcode := range opcodes {
		if operand >= len(opcode.Operands) {
			panic(fmt.Sprintf("Opcode %s expects only %d operands", opcode.String(), len(opcode.Operands)))
		}
		if opcode.Operands[operand].Type == OT_rel8 {
			opcodeMap.add(lib.T_Uint8, lib.BYTE, opcode)
		} else if opcode.Operands[operand].Type == OT_rel16 {
			opcodeMap.add(lib.T_Uint16, lib.WORD, opcode)
		} else if opcode.Operands[operand].Type == OT_rel32 {
			opcodeMap.add(lib.T_Uint32, lib.DOUBLE, opcode)
		} else if opcode.Operands[operand].Type == OT_rm8 {
			opcodeMap.add(lib.T_Register, lib.BYTE, opcode)
			opcodeMap.add(lib.T_IndirectRegister, lib.BYTE, opcode)
			opcodeMap.add(lib.T_DisplacedRegister, lib.BYTE, opcode)
			opcodeMap.add(lib.T_RIPRelative, lib.BYTE, opcode)
			opcodeMap.add(lib.T_SIBRegister, lib.QUADWORD, opcode)
		} else if opcode.Operands[operand].Type == OT_rm16 {
			opcodeMap.add(lib.T_Register, lib.WORD, opcode)
			opcodeMap.add(lib.T_IndirectRegister, lib.WORD, opcode)
			opcodeMap.add(lib.T_DisplacedRegister, lib.WORD, opcode)
			opcodeMap.add(lib.T_RIPRelative, lib.WORD, opcode)
			opcodeMap.add(lib.T_SIBRegister, lib.QUADWORD, opcode)
		} else if opcode.Operands[operand].Type == OT_rm32 {
			opcodeMap.add(lib.T_Register, lib.DOUBLE, opcode)
			opcodeMap.add(lib.T_IndirectRegister, lib.DOUBLE, opcode)
			opcodeMap.add(lib.T_DisplacedRegister, lib.DOUBLE, opcode)
			opcodeMap.add(lib.T_RIPRelative, lib.DOUBLE, opcode)
			opcodeMap.add(lib.T_SIBRegister, lib.QUADWORD, opcode)
		} else if opcode.Operands[operand].Type == OT_rm64 {
			opcodeMap.add(lib.T_Register, lib.QUADWORD, opcode)
			opcodeMap.add(lib.T_IndirectRegister, lib.QUADWORD, opcode)
			opcodeMap.add(lib.T_DisplacedRegister, lib.QUADWORD, opcode)
			opcodeMap.add(lib.T_RIPRelative, lib.QUADWORD, opcode)
			opcodeMap.add(lib.T_SIBRegister, lib.QUADWORD, opcode)
		} else if opcode.Operands[operand].Type == OT_m {
			opcodeMap.add(lib.T_DisplacedRegister, lib.QUADWORD, opcode)
			opcodeMap.add(lib.T_RIPRelative, lib.QUADWORD, opcode)
		} else if opcode.Operands[operand].Type == OT_m16 {
			opcodeMap.add(lib.T_IndirectRegister, lib.WORD, opcode)
		} else if opcode.Operands[operand].Type == OT_m32 {
			opcodeMap.add(lib.T_IndirectRegister, lib.DOUBLE, opcode)
		} else if opcode.Operands[operand].Type == OT_m64 {
			opcodeMap.add(lib.T_IndirectRegister, lib.QUADWORD, opcode)
		} else if opcode.Operands[operand].Type == OT_imm8 {
			opcodeMap.add(lib.T_Uint8, lib.BYTE, opcode)
		} else if opcode.Operands[operand].Type == OT_imm16 {
			opcodeMap.add(lib.T_Uint16, lib.WORD, opcode)
		} else if opcode.Operands[operand].Type == OT_imm32 {
			opcodeMap.add(lib.T_Uint32, lib.DOUBLE, opcode)
		} else if opcode.Operands[operand].Type == OT_imm64 {
			opcodeMap.add(lib.T_Uint64, lib.QUADWORD, opcode)
			opcodeMap.add(lib.T_Float64, lib.QUADWORD, opcode)
		} else if opcode.Operands[operand].Type == OT_r8 {
			opcodeMap.add(lib.T_Register, lib.BYTE, opcode)
		} else if opcode.Operands[operand].Type == OT_r16 {
			opcodeMap.add(lib.T_Register, lib.WORD, opcode)
		} else if opcode.Operands[operand].Type == OT_r32 {
			opcodeMap.add(lib.T_Register, lib.DOUBLE, opcode)
		} else if opcode.Operands[operand].Type == OT_r64 {
			opcodeMap.add(lib.T_Register, lib.QUADWORD, opcode)
		} else if opcode.Operands[operand].Type == OT_xmm1 {
			opcodeMap.add(lib.T_Register, lib.OWORD, opcode)
		} else if opcode.Operands[operand].Type == OT_xmm2 {
			opcodeMap.add(lib.T_Register, lib.OWORD, opcode)
		} else if opcode.Operands[operand].Type == OT_xmm1m64 {
			opcodeMap.add(lib.T_Register, lib.OWORD, opcode)
			opcodeMap.add(lib.T_Register, lib.QUADWORD, opcode)
			opcodeMap.add(lib.T_IndirectRegister, lib.QUADWORD, opcode)
			opcodeMap.add(lib.T_RIPRelative, lib.QUADWORD, opcode)
			opcodeMap.add(lib.T_SIBRegister, lib.QUADWORD, opcode)
		} else if opcode.Operands[operand].Type == OT_xmm2m64 {
			opcodeMap.add(lib.T_Register, lib.OWORD, opcode)
			opcodeMap.add(lib.T_Register, lib.QUADWORD, opcode)
			opcodeMap.add(lib.T_IndirectRegister, lib.QUADWORD, opcode)
			opcodeMap.add(lib.T_RIPRelative, lib.QUADWORD, opcode)
			opcodeMap.add(lib.T_SIBRegister, lib.QUADWORD, opcode)
		} else if opcode.Operands[operand].Type == OT_xmm2m128 {
			opcodeMap.add(lib.T_Register, lib.OWORD, opcode)
			opcodeMap.add(lib.T_Register, lib.QUADWORD, opcode)
			opcodeMap.add(lib.T_RIPRelative, lib.QUADWORD, opcode)
		} else if opcode.Operands[operand].Type == OT_ymm1 {
			opcodeMap.add(lib.T_Register, lib.YWORD, opcode)
		} else if opcode.Operands[operand].Type == OT_ymm2 {
			opcodeMap.add(lib.T_Register, lib.YWORD, opcode)
		}
	}
	return opcodeMap
}
