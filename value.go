package pmmp

import (
    "fmt"
    "math"

    "github.com/npillmayer/arithm"
    "github.com/npillmayer/arithm/polyn"
    "github.com/npillmayer/schuko/tracing"
)

// tracer traces with key 'pmmp'.
func tracer() tracing.Trace {
    return tracing.Select("pmmp")
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

// Value is an interface for all values which PMMP can handle.
type Value interface {
    Self() ValueBase // helper indirection, see type ValueBase
    IsKnown() bool   // is this a known value ?
    Type() ValueType // type of the value
}

// ValueBase is a helper struct for operations on values.
type ValueBase struct {
    V Value
}

func (b ValueBase) String() string {
    if b.IsNumeric() {
        return b.AsNumeric().Polynomial().String()
    } else if b.IsPair() {
        p := b.AsPair()
        return "(" + p.xpart.Self().String() + "," + p.ypart.Self().String() + ")"
    }
    return fmt.Sprintf("%v", b.V)
}

// IsNumeric is a predicate: is it a Numeric?
func (b ValueBase) IsNumeric() bool {
    _, ok := b.V.(Numeric)
    return ok
}

// IsPair is a predicate: is it a Pair?
func (b ValueBase) IsPair() bool {
    _, ok := b.V.(Pair)
    return ok
}

// Type returns the value type of a value.
func (b ValueBase) Type() ValueType {
    return b.V.Type()
}

// AsNumeric returns a value a Numeric, or an error and an invalid Numeric.
func (b ValueBase) AsNumeric() Numeric {
    if n, ok := b.V.(Numeric); ok {
        return n
    }
    tracer().Errorf("value is not of type numeric: %v", b.V)
    return Numeric{}
}

// AsPair returns a value as a Pair, or an error and a NullPair.
func (b ValueBase) AsPair() Pair {
    if p, ok := b.V.(Pair); ok {
        return p
    }
    tracer().Errorf("value is not of type numeric: %v", b.V)
    return NullPair()
}

// Minus calculates a - b.
func (b ValueBase) Minus(w Value) (Value, error) {
    switch b.Type() {
    case NumericType:
        if w.Self().IsNumeric() {
            return b.AsNumeric().Minus(w.Self().AsNumeric()), nil
        }
    case PairType:
        if w.Self().IsPair() {
            return b.AsPair().Minus(w.Self().AsPair()), nil
        }
    default:
        tracer().Errorf("not yet implemented: %T minus %T", b.V, w)
    }
    return Numeric{}, fmt.Errorf("not yet implemented: %T minus %T", b.V, w)
}

// --- Numeric ---------------------------------------------------------------

// Numeric is a known or unknown scalar value.
type Numeric polyn.Polynomial

// Self returns this numeric, wrapped into a ValueBase struct.
func (n Numeric) Self() ValueBase {
    return ValueBase{n}
}

// Polynomial is a type-cast to type Polynomial.
func (n Numeric) Polynomial() polyn.Polynomial {
    return polyn.Polynomial(n)
}

// IsKnown is a predicate: is this a known value?
func (n Numeric) IsKnown() bool {
    if !polyn.Polynomial(n).IsValid() {
        return false
    }
    _, ok := polyn.Polynomial(n).IsConstant()
    return ok
}

// Type returns NumericType.
func (n Numeric) Type() ValueType {
    return NumericType
}

// FromFloat creates a numeric value from a float.
func FromFloat(f float64) Numeric {
    if math.IsNaN(f) {
        panic("cannot create numeric value from NaN")
    }
    return Numeric(polyn.NewConstantPolynomial(f))
}

// AsFloat returns a known numeric value as a float, or NaN.
func (n Numeric) AsFloat() float64 {
    if !n.IsKnown() {
        tracer().Errorf("value is not a known numeric: %v", n)
        return math.NaN()
    }
    return polyn.Polynomial(n).GetConstantValue()
}

// Plus is n + m.
func (n Numeric) Plus(m Numeric) Numeric {
    r := polyn.Polynomial(n).Add(polyn.Polynomial(m), false)
    return Numeric(r)
}

// Minus is n - m.
func (n Numeric) Minus(m Numeric) Numeric {
    r := polyn.Polynomial(n).Subtract(polyn.Polynomial(m), false)
    return Numeric(r)
}

// --- Pair ------------------------------------------------------------------

// Pair is a known or unknown pair value.
type Pair struct {
    xpart Numeric
    ypart Numeric
}

// NullPair is an invalid pair.
func NullPair() Pair {
    return Pair{
        xpart: Numeric{},
        ypart: Numeric{},
    }
}

// NewPair creates a new pair from two scalar values.
func NewPair(a, b Numeric) Pair {
    return Pair{
        xpart: a,
        ypart: b,
    }
}

// Self returns this pair, wrapped into a ValueBase struct.
func (p Pair) Self() ValueBase {
    return ValueBase{p}
}

// IsKnown is a predicate: is this a known value?
func (p Pair) IsKnown() bool {
    return p.xpart.IsKnown() && p.ypart.IsKnown()
}

// Type returns PairType
func (p Pair) Type() ValueType {
    return PairType
}

// XNumeric returns the scalar xpart.
func (p Pair) XNumeric() Numeric {
    return p.xpart
}

// YNumeric returns the scalar ypart.
func (p Pair) YNumeric() Numeric {
    return p.ypart
}

// AsPair returns a known pair value.
func (p Pair) AsPair() arithm.Pair {
    if !p.IsKnown() {
        tracer().Errorf("value is not a known pair: %v", p)
        var none arithm.Pair
        return none
    }
    return arithm.P(p.xpart.AsFloat(), p.ypart.AsFloat())
}

// ConvPair converts a pair to a pair value.
func ConvPair(p arithm.Pair) Pair {
    return NewPair(
        FromFloat(p.X()),
        FromFloat(p.Y()),
    )
}

// Minus is p - q.
func (p Pair) Minus(q Pair) Pair {
    r := NewPair(
        p.XNumeric().Minus(q.XNumeric()),
        p.YNumeric().Minus(q.YNumeric()),
    )
    return r
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
