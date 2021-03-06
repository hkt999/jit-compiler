package x86_64

import (
	"github.com/bspaans/jit-compiler/asm/x86_64"
	"github.com/bspaans/jit-compiler/asm/x86_64/encoding"
	"github.com/bspaans/jit-compiler/ir/expr"
	. "github.com/bspaans/jit-compiler/ir/shared"
	"github.com/bspaans/jit-compiler/lib"
)

func encode_IR_Int8(i *expr.IR_Int8, ctx *IR_Context, target lib.Operand) ([]lib.Instruction, error) {
	result := []lib.Instruction{x86_64.MOV(encoding.Uint8(int8(i.Value)), target)}
	ctx.AddInstruction(result...)
	return result, nil
}
