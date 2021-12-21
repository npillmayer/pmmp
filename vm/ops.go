package vm

import "fmt"

type OpCode uint16

//go:generate stringer -type OpCode
const (
	OpNop OpCode = 0

	OpArgI OpCode = 1 // int argument
	OpArgF OpCode = 2 // float argument
	OpArgP OpCode = 3 // pair argument
	OpArgV OpCode = 4 // variable ref argument
	OpArgS OpCode = 5 // string argument

	Const OpCode = (8 + iota) << 3 // FCONST ⟪f64⟫ : put a float constant onto the stack

	IConst OpCode = Const + OpArgI
	FConst OpCode = Const + OpArgF
)

type Op struct {
	opcode OpCode
	arg    interface{}
}

type RegisterSet struct {
	I int
	F float64
	S string
}

func (rset *RegisterSet) DecodeArg(op Op) {
	switch op.opcode & 0x07 {
	case 1:
		rset.I = op.arg.(int)
	case 2:
		rset.F = op.arg.(float64)
	case 5:
		rset.S = op.arg.(string)
	default:
		panic(fmt.Sprintf("TODO cannot yet handle op %2x", op.opcode))
	}
}
