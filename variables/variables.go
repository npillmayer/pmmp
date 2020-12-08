package variables

import (
	"bytes"
	"fmt"
	"math"
	"strings"

	"github.com/npillmayer/arithm"
	"github.com/npillmayer/gorgo/runtime"
	"github.com/npillmayer/schuko/gtrace"
	"github.com/npillmayer/schuko/tracing"
)

// T traces to the global syntax tracer
func T() tracing.Trace {
	return gtrace.SyntaxTracer
}

// === Variable Type Declarations ============================================

// VariableType obviously represents the type of a variable.
type VariableType int8

// Predefined variable types
const (
	Undefined VariableType = iota
	NumericType
	PairType
	PathType
	ColorType
	PenType
	VardefType
	SubscriptType
	SuffixType
)

/*
VarDecl represents a variable declaration.

MetaFont declares variables explicitly (`numeric x`) or dynamically
(`x=1` ⟹ x is of type numeric). Dynamic typing is permitted for
numeric variables only, all other types must be declared. In our
implementation, declaration
is for tags only, i.e. the `x` in `x2r`. This differs from MetaFont,
where x2r can have a separate type from x.

We build up a doubly-linked tree of variable declarations to describe a
variable with a single defining tag. The tag is the entity that goes to
the symbol table (of a scope). Suffixes and subscripts are attached to
the tag, but invisible as symbols.

Example:

   numeric x; x2r := 7; x.b := 77;

Result:

  tag = "x"  of type NumericType ⟹ into symbol table of a scope
     +-- suffix ".b" of type SuffixType           "x.b"
     +-- subscript "[]" of type SubscriptType:    "x[]"
         +--- suffix ".r" of type SuffixType:     "x[].r"

*/
type VarDecl struct { // this is a tag, an array-subtype, or a suffix
	Suffix
	tag *runtime.Tag // declared tag
}

// Suffix is a variable suffix (subscript or tag)
type Suffix struct {
	name        string
	isSubscript bool
	Parent      *Suffix
	Sibling     *Suffix
	Suffixes    *Suffix
	baseDecl    *VarDecl
}

// NewVarDecl creates and initialize a new variable type declaration.
func NewVarDecl(nm string, tp VariableType) *VarDecl {
	decl := &VarDecl{}
	decl.name = nm
	decl.tag = runtime.NewTag(nm).WithType(int8(tp))
	decl.tag.UData = decl
	decl.baseDecl = decl // this pointer should never be nil
	T().P("decl", decl.FullName()).Debugf("atomic variable type declaration created")
	return decl
}

// Tag returns the variable declaration as a tag (for symbol tables).
func (d *VarDecl) Tag() *runtime.Tag {
	return d.tag
}

func varDeclFromTag(tag *runtime.Tag) *VarDecl {
	return tag.UData.(*VarDecl)
}

// AsSuffix returns a variable declaration as a Suffix. Every variable
// declaration is also a suffix, although one without a parent.
func (d *VarDecl) AsSuffix() *Suffix {
	return &d.Suffix
}

func (d *VarDecl) String() string {
	return fmt.Sprintf("<decl %s/%s>", d.FullName(), d.GetBaseType().String())
}

// IsBaseDecl is a prediacte: is this suffix the base tag?
func (s *Suffix) IsBaseDecl() bool {
	return s.Parent == nil
}

// Type returns the variable's type.
func (s *Suffix) Type() VariableType {
	return VariableType(s.baseDecl.tag.Type())
}

// Name gets the isolated name of declaration partial (tag, array or suffix).
func (s *Suffix) Name() string {
	if s.isSubscript {
		return "[]"
	} else if s.Parent != nil {
		return "." + s.name
	}
	return s.name
}

// FullName gets
// the full name of a type declaration, starting with the base tag.
// x <- array <- suffix(a)  gives "x[].a", which is a bit more verbose
// than MetaFont's response. I prefer this one.
//
func (s *Suffix) FullName() string {
	if s.Parent == nil {
		return s.baseDecl.tag.Name
	} // we are in a declaration for a complex type
	var str bytes.Buffer
	str.WriteString(s.Parent.FullName()) // recursive
	if s.isSubscript {
		str.WriteString("[]")
	} else {
		str.WriteString(".") // MetaFont suppresses this if following a subscript partial
		str.WriteString(s.name)
	}
	return str.String()
}

// GetBaseType gets the type of the base tag.
func (s *Suffix) GetBaseType() VariableType {
	return s.baseDecl.Type()
}

// CreateSuffix is a function to create and initialize a suffix to
// a type declaration.
// Callers provide a (usually complex) type and an optional parent.
// If the parent is given and already has a child / suffix-partial with
// the same signature as the one to create, this function will not create
// a new partial, but provide the existing one.
//
// Will panic if tp is neither SuffixType nor SubscriptType (use
// NewVarDecl() instead).
//
// Will panic if SubscriptType, but no parent given.
//
func CreateSuffix(nm string, tp VariableType, parent *Suffix) *Suffix {
	if tp != SubscriptType && tp != SuffixType {
		panic("Suffix must either be of type SuffixType or SubscriptType")
	}
	if parent != nil { // check if already exists as child of parent
		if parent.Suffixes != nil {
			ch := parent.Suffixes
			for ch != nil { // check all direct suffixes
				if (tp == SuffixType && ch.name == nm) ||
					(ch.isSubscript && tp == SubscriptType) {
					T().P("decl", ch.FullName()).Debugf("variable type already declared")
					return ch // we're done
				}
				ch = ch.Sibling
			}
		}
	}
	s := &Suffix{ // not found => create a new one
		name:        nm,
		isSubscript: tp == SubscriptType,
	}
	if parent == nil { // then wrap into a var decl
		if s.isSubscript { // can't have a subscript as a top-level var decl
			panic("Trying to have standalone subscript")
		}
		d := NewVarDecl(nm, NumericType) // numeric is default in MetaFont
		d.Suffix = *s
		s.baseDecl = d
		return d.AsSuffix()
	}
	s.appendToSuffix(parent) // parent given => append to it
	T().P("decl", s.FullName()).Debugf("variable type decl suffix created")
	return s
}

// appendToSuffix appends a complex type partial (suffix or array) to a parent identifier.
// Will not append the partial, if a partial with this name already
// exists (as a child).
//
func (s *Suffix) appendToSuffix(parent *Suffix) *Suffix {
	if parent == nil {
		panic("attempt to append type declaration to nil-tag")
	}
	s.baseDecl = parent.baseDecl
	s.Parent = parent
	ch := parent.Suffixes
	if ch == nil {
		parent.Suffixes = s
	} else {
		for ch.Sibling != nil {
			ch = ch.Sibling
		}
		ch.Sibling = s
	}
	T().P("decl", s.FullName()).Debugf("new variable type suffix declaration")
	return s
}

// ShowDeclarations shows the variable declarations tree, usually starting at the
// base tag (MetaFont 'show' command).
func (s *Suffix) ShowDeclarations(b *bytes.Buffer) *bytes.Buffer {
	if b == nil {
		b = new(bytes.Buffer)
	}
	b.WriteString(fmt.Sprintf("%s : %s\n", s.FullName(), s.baseDecl.GetBaseType().String()))
	ch := s.Suffixes
	for ; ch != nil; ch = ch.Sibling {
		b = ch.ShowDeclarations(b)
	}
	return b
}

// === Variable References / Usage ===========================================

// VarRef is a varibale reference.
// Variable reference look like "x", "x2", "hello.world" or "a[4.32].b".
// Variable references always refer to variable declarations (see
// type VarDecl), which define the type and structure of the variable.
//
// The declaration may have partials of type subscript. For every such
// partial the reference needs a decimal subscript, which we will store
// in an array of subscripts.
//
// Example:
//
//     x[2.8].b[1] => subscripts = [2.8, 1]
//
// Variable references can have a value (of type interface{}).
//
// All in all, variable references may get pretty complicated. This is due
// to the complicated nature of MetaFont/-Post variables, but also from
// the fact, that variable values must be solvable by a system of linear
// equations (LEQ). Having variable values described by a LEQ is a core
// feature of MetaPost, so we better invest some effort.
//
type VarRef struct {
	runtime.Tag             // store by normalized name
	cachedName  string      // store full name
	decl        *Suffix     // type declaration for this variable
	subscripts  []float64   // list of subscript value, first to last
	Value       interface{} // if known: has a value (numeric, pair, path, ...)
}

// CreateVarRef creates a variable reference. Low level method.
func CreateVarRef(decl *Suffix, value interface{}, indices []float64) *VarRef {
	if decl.GetBaseType() == PairType {
		return CreatePairTypeVarRef(decl, value, indices)
	}
	T().Debugf("creating %s var for %v", decl.Type().String(), decl)
	v := &VarRef{
		decl:       decl,
		subscripts: indices,
		Value:      value,
	}
	v.ChangeType(int8(decl.Type()))
	v.ID = newVarSerial() // TODO: check, when this is needed (now: id leak)
	//T().Debugf("created var ref: subscripts = %v", indices)
	v.Tag.UData = v // link back to var from tag
	return v
}

func (v *VarRef) String() string {
	return fmt.Sprintf("<var %s=%v w/ %s>", v.FullName(), v.Value, v.decl.FullName())
}

// Name returns the full nomalized name, i.e. "x[2].r".
// This enables us to store the variable in a symbol table.
// Overrides Name of interface Symbol.
//
func (v *VarRef) Name() string {
	if len(v.cachedName) == 0 {
		v.cachedName = v.FullName()
	}
	return v.cachedName
}

// Type returns the variable's type.
func (v *VarRef) Type() VariableType {
	if v.decl != nil {
		return v.decl.GetBaseType()
	}
	return Undefined
}

func varFromTag(tag *runtime.Tag) *VarRef {
	return tag.UData.(*VarRef)
}

// GetSuffixesString returns a variable's name without the leading base tag name.
// Strip the base tag string off of a variable and return all the suffxies
// as string.
//
// func (v *VarRef) GetSuffixesString() string {
// 	basetag := v.Decl.baseDecl
// 	basetagstring := basetag.Name
// 	fullstring := v.FullName()
// 	return fullstring[len(basetagstring):]
// }

// --- Variable Type Pair ----------------------------------------------------

// PairPartRef are either xpart or ypart of a pair.
//
// Complications for pairs arise because we need each half of the pair
// (x-part and y-part) as separate variables for the LEQ-solver.
// Variables of type pair will use two sub-symbols for the x-part and
// y-part of the pair respectively. We will connect them using the
// sibling-link (x-part) and child-link (y-part) of the VarRef.
// Both parts link back to the pair variable.
//
// We need a different serial ID for the y-part, as it will be used as a
// variable index in a system of linear equations LEQ. Otherwise x-part and
// y-part would not be distinguishable for the LEQ.
//
type PairPartRef struct {
	runtime.Tag
	ID      int32       // serial ID
	Pairvar *VarRef     // pair parent
	Value   interface{} // if known: has a value (numeric)
}

// CreatePairTypeVarRef creates a pair variable reference. Low level method.
func CreatePairTypeVarRef(decl *Suffix, value interface{}, indices []float64) *VarRef {
	T().Debugf("creating pair var for %v", decl.baseDecl.FullName())
	v := &VarRef{
		decl:       decl,
		subscripts: indices,
		Value:      value,
	}
	v.Tag.Name = decl.name
	v.ChangeType(int8(PairType))
	v.ID = newVarSerial() // TODO: check, when this is needed (now: id leak)
	var pair arithm.Pair
	var ok bool
	ypart := &PairPartRef{
		ID:      newVarSerial(),
		Pairvar: v,
	}
	ypart.Tag.UData = ypart
	xpart := &PairPartRef{
		ID:      v.ID,
		Pairvar: v,
	}
	xpart.Tag.UData = xpart
	v.Sibling = &xpart.Tag
	v.Children = &ypart.Tag
	if pair, ok = value.(arithm.Pair); ok {
		T().Debugf("setting value of pair var to %v", pair)
		xpart.Value = pair.X()
		ypart.Value = pair.Y()
	} else {
		xpart.Value = math.NaN()
		ypart.Value = math.NaN()
	}
	return v
}

// Name returns a string for a pair part.
// Pair parts (x-part or y-part) return the name of their parent pair symbol,
// prepending "xpart" or "ypart" respectively. This name is constant and
// may be used to store the pair part in a symbol table.
//
func (ppart *PairPartRef) Name() string {
	if ppart.Pairvar.ID == ppart.ID {
		return "xpart " + ppart.Pairvar.Tag.Name
	}
	return "ypart " + ppart.Pairvar.Tag.Name
}

// Type returns the type of a pair part, which is always numeric.
func (ppart *PairPartRef) Type() VariableType {
	return NumericType
}

func ppFromTag(tag *runtime.Tag) *PairPartRef {
	return tag.UData.(*PairPartRef)
}

func (ppart *PairPartRef) String() string {
	return "<" + ppart.Name() + fmt.Sprintf("=%v>", ppart.Value.(float64))
}

// IsPair is a predicate: is this variable of type pair?
func (v *VarRef) IsPair() bool {
	return v.Type() == PairType
}

// XPart gets the x-part of a pair variable
func (v *VarRef) XPart() *PairPartRef {
	if !v.IsPair() {
		T().P("var", v.Name).Errorf("cannot access x-part of non-pair")
		return nil
	}
	return ppFromTag(v.Sibling)
}

// YPart gets the y-part of a pair variable
func (v *VarRef) YPart() *PairPartRef {
	if !v.IsPair() {
		T().P("var", v.Name).Errorf("cannot access y-part of non-pair")
		return nil
	}
	return ppFromTag(v.Children)
}

/*
Get the x-part value of a pair.

Interface runtime.Assignable
*/
// func (ppart *PairPartRef) GetValue() interface{} {
// 	return ppart.Value
// }

// Interface runtime.Assignable
// func (ppart *PairPartRef) SetValue(val interface{}) {
// 	T().P("var", ppart.Name).Debugf("new value: %v", val)
// 	ppart.Value = val
// 	ppart.Pairvar.PullValue()
// }

// IsKnown returns wether this pair part variable has a known value.
func (ppart *PairPartRef) IsKnown() bool {
	return (ppart.Value != nil)
}

// FullName gets the full normalized (canonical) name of a variable,  i.e.
//
//    "x[2].r".
//
func (v *VarRef) FullName() string {
	var suffixes []string
	d := v.decl
	if d == nil {
		return fmt.Sprintf("<undeclared variable: %s>", v.Name())
	}
	subscriptcount := len(v.subscripts) - 1
	for sfx := v.decl; sfx != nil; sfx = sfx.Parent { // iterate backwards
		//T().Printf("sfx = %v", sfx)
		//if sfx.Type() == SubscriptType {
		if sfx.isSubscript {
			s := "[" + fmt.Sprintf("%g", v.subscripts[subscriptcount]) + "]"
			suffixes = append(suffixes, s)
			subscriptcount--
		} else {
			suffixes = append(suffixes, sfx.Name())
		}
	}
	// now reverse the slice of suffixes
	for i := 0; i < len(suffixes)/2; i++ { // swap suffixes in place
		j := len(suffixes) - i - 1
		suffixes[i], suffixes[j] = suffixes[j], suffixes[i]
	}
	return strings.Join(suffixes, "")
}

// GetValue satisfies interface runtime.Assignable
func (v *VarRef) GetValue() interface{} {
	return v.Value
}

// SetValue satisfies interface runtime.Assignable
func (v *VarRef) SetValue(val interface{}) {
	T().P("var", v.Name).Debugf("new value: %v", val)
	v.Value = val
	if v.IsPair() {
		var xpart *PairPartRef = ppFromTag(v.Sibling)
		var ypart *PairPartRef = ppFromTag(v.Children)
		if val == nil {
			xpart.Value = nil
			ypart.Value = nil
		} else {
			var pairval arithm.Pair = val.(arithm.Pair)
			xpart.Value = pairval.X
			ypart.Value = pairval.Y
		}
	}
}

// PullValue pulls a variable's value up to a containig pair variable.
// Whenever a pair part (x-part or y-part) is set, it sends a message to
// the parent pair variable to pull the value. If both parts are known and
// a numeric value is set, the parent pair creates a combined pair value.
//
func (v *VarRef) PullValue() {
	if v.IsPair() {
		var ppart1, ppart2 *PairPartRef
		ppart1 = ppFromTag(v.Sibling)
		ppart2 = ppFromTag(v.Children)
		if ppart1 != nil && ppart2 != nil {
			if ppart1.Value != nil && ppart2.Value != nil {
				v.Value = arithm.P(ppart1.Value.(float64), ppart2.Value.(float64))
				T().P("var", v.Name).Debugf("pair value = %s",
					v.Value.(arithm.Pair).String())
			}
		}
	}
}

// ValueString gets
// the value of a variable as a string, if known. Otherwise return
// the tag name or type, depending on the variable type.
func (v *VarRef) ValueString() string {
	if v.IsPair() {
		var xvalue, yvalue string
		xpart := v.XPart().Value
		if xpart == nil {
			xvalue = fmt.Sprintf("xpart %s", v.Name())
		} else {
			xvalue = fmt.Sprintf("%g", xpart.(float64))
		}
		ypart := v.YPart().Value
		if ypart == nil {
			yvalue = fmt.Sprintf("ypart %s", v.Name())
		} else {
			yvalue = fmt.Sprintf("%g", ypart.(float64))
		}
		return fmt.Sprintf("(%s,%s)", xvalue, yvalue)
	}
	if v.IsKnown() {
		return fmt.Sprintf("%v", v)
	}
	return "<numeric>"
}

// IsKnown returns wether this variable has a known value.
func (v *VarRef) IsKnown() bool {
	return (v.Value != nil)
}

// Reincarnate sets a
// new ID for a variable reference. Whenever variables become
// re-incarnated, a new serial ID is needed. Re-incarnation happens,
// whenever a variable goes out of scope, but is still relevant in the
// LEQ-system. The variables' name continues to live on in a new incarnation,
// while the out-of-scope variable lives on with the old serial.
//
// Returns the old serial ID.
//
func (v *VarRef) Reincarnate() int32 {
	oldserial := v.ID
	v.ID = newVarSerial()
	if v.IsPair() {
		ypartid := newVarSerial()
		v.XPart().ID = v.ID
		v.YPart().ID = ypartid
	}
	return oldserial
}

var varSerial int32 = 1 // serial number counter for variables, always > 0 !

// Get a new unique serial ID for variables.
func newVarSerial() int32 {
	serial := varSerial
	varSerial++
	T().Debugf("creating new serial ID %d", serial)
	return serial
}

// --- Helpers ---------------------------------------------------------------

// TypeString returns a type as string.
func (vt VariableType) String() string {
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
func TypeFromString(str string) VariableType {
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
