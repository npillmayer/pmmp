// Code generated by "stringer -type tokType"; DO NOT EDIT.

package grammar

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[String - -1]
	_ = x[Ident - -2]
	_ = x[Literal - -3]
	_ = x[SymTok - -9]
	_ = x[Unsigned - -10]
	_ = x[Signed - -11]
	_ = x[UnaryOp - -15]
	_ = x[NullaryOp - -16]
	_ = x[PrimaryOp - -17]
	_ = x[SecondaryOp - -18]
	_ = x[RelationOp - -19]
	_ = x[AssignOp - -20]
	_ = x[OfOp - -21]
	_ = x[UnaryTransform - -22]
	_ = x[BinaryTransform - -23]
	_ = x[PlusOrMinus - -24]
	_ = x[Type - -25]
	_ = x[PseudoOp - -26]
	_ = x[Function - -27]
	_ = x[Join - -28]
	_ = x[DrawCmd - -29]
	_ = x[DrawOption - -30]
	_ = x[ScalarMulOp - -31]
	_ = x[MacroDef - -32]
	_ = x[Keyword - -33]
}

const (
	_tokType_name_0 = "KeywordMacroDefScalarMulOpDrawOptionDrawCmdJoinFunctionPseudoOpTypePlusOrMinusBinaryTransformUnaryTransformOfOpAssignOpRelationOpSecondaryOpPrimaryOpNullaryOpUnaryOp"
	_tokType_name_1 = "SignedUnsignedSymTok"
	_tokType_name_2 = "LiteralIdentString"
)

var (
	_tokType_index_0 = [...]uint8{0, 7, 15, 26, 36, 43, 47, 55, 63, 67, 78, 93, 107, 111, 119, 129, 140, 149, 158, 165}
	_tokType_index_1 = [...]uint8{0, 6, 14, 20}
	_tokType_index_2 = [...]uint8{0, 7, 12, 18}
)

func (i tokType) String() string {
	switch {
	case -33 <= i && i <= -15:
		i -= -33
		return _tokType_name_0[_tokType_index_0[i]:_tokType_index_0[i+1]]
	case -11 <= i && i <= -9:
		i -= -11
		return _tokType_name_1[_tokType_index_1[i]:_tokType_index_1[i+1]]
	case -3 <= i && i <= -1:
		i -= -3
		return _tokType_name_2[_tokType_index_2[i]:_tokType_index_2[i+1]]
	default:
		return "tokType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}
