package pmmp

import (
	"fmt"

	"github.com/npillmayer/arithm"
	"github.com/npillmayer/arithm/polyn"
	"github.com/npillmayer/schuko/gtrace"
	"github.com/npillmayer/schuko/tracing"
)

// T traces to the global syntax tracer.
func T() tracing.Trace {
	return gtrace.SyntaxTracer
}

// ValueType represents the type of a value.
type ValueType int8

// Predefined variable types
const (
	Undefined ValueType = iota
	NumericType
	PairType
	PathType
	ColorType
	PenType
	VardefType
	SubscriptType
	SuffixType
)

// --- Value -----------------------------------------------------------------

type Value interface {
	Self() ValueBase
	IsKnown() bool
	Type() ValueType
}

type ValueBase struct {
	V Value
}

func (b ValueBase) IsNumeric() bool {
	_, ok := b.V.(Numeric)
	return ok
}

func (b ValueBase) IsPair() bool {
	_, ok := b.V.(Pair)
	return ok
}

func (b ValueBase) Type() ValueType {
	return b.V.Type()
}

func (b ValueBase) AsNumeric() Numeric {
	if n, ok := b.V.(Numeric); ok {
		return n
	}
	T().Errorf("value is not of type numeric: %v", b.V)
	return Numeric{}
}

func (b ValueBase) AsPair() Pair {
	if p, ok := b.V.(Pair); ok {
		return p
	}
	T().Errorf("value is not of type numeric: %v", b.V)
	return NullPair()
}

// --- Numeric ---------------------------------------------------------------

type Numeric polyn.Polynomial

func (n Numeric) Self() ValueBase {
	return ValueBase{n}
}

func (n Numeric) IsKnown() bool {
	if !polyn.Polynomial(n).IsValid() {
		return false
	}
	_, ok := polyn.Polynomial(n).IsConstant()
	return ok
}

func (n Numeric) Type() ValueType {
	return NumericType
}

func FromFloat(f float64) Numeric {
	return Numeric(polyn.NewConstantPolynomial(f))
}

func (n Numeric) AsFloat() float64 {
	if !n.IsKnown() {
		T().Errorf("value is not a known numeric: %v", n)
		return 0
	}
	return polyn.Polynomial(n).GetConstantValue()
}

// --- Pair ------------------------------------------------------------------

type Pair struct {
	xpart Numeric
	ypart Numeric
}

func NullPair() Pair {
	return Pair{
		xpart: Numeric{},
		ypart: Numeric{},
	}
}

func NewPair(a, b Numeric) Pair {
	return Pair{
		xpart: a,
		ypart: b,
	}
}

func (p Pair) Self() ValueBase {
	return ValueBase{p}
}

func (p Pair) IsKnown() bool {
	return p.xpart.IsKnown() && p.ypart.IsKnown()
}

func (p Pair) Type() ValueType {
	return PairType
}

func (p Pair) XNumeric() Numeric {
	return p.xpart
}

func (p Pair) YNumeric() Numeric {
	return p.ypart
}

func (p Pair) AsPair() arithm.Pair {
	if !p.IsKnown() {
		T().Errorf("value is not a known pair: %v", p)
		var none arithm.Pair
		return none
	}
	return arithm.P(p.xpart.AsFloat(), p.ypart.AsFloat())
}

func ConvPair(p arithm.Pair) Pair {
	return NewPair(
		FromFloat(p.X()),
		FromFloat(p.Y()),
	)
}

// --- Helpers ---------------------------------------------------------------

func (vt ValueType) String() string {
	switch vt {
	case Undefined:
		return "<undefined>"
	case NumericType:
		return "numeric"
	case PairType:
		return "pair"
	case PathType:
		return "path"
	case ColorType:
		return "color"
	case PenType:
		return "pen"
	case VardefType:
		return "vardef"
	case SubscriptType:
		return "[]"
	case SuffixType:
		return "<suffix>"
	}
	return fmt.Sprintf("<illegal type: %d>", vt)
}

// TypeFromString gets a type from a string.
func TypeFromString(str string) ValueType {
	switch str {
	case "numeric":
		return NumericType
	case "pair":
		return PairType
	case "path":
		return PathType
	case "color":
		return ColorType
	case "pen":
		return PenType
	}
	return Undefined
}
