package variables

import (
	"bytes"
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/npillmayer/gorgo/runtime"
	"github.com/npillmayer/pmmp"
	"github.com/npillmayer/schuko/gtrace"
	"github.com/npillmayer/schuko/tracing"
)

// T traces to the global syntax tracer
func T() tracing.Trace {
	return gtrace.SyntaxTracer
}

// === Variable Type Declarations ============================================

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
	runtime.Tag // declared tag
	Suffix
}

// Suffix is a variable suffix (subscript or tag)
type Suffix struct {
	suffixName  string
	isSubscript bool
	Parent      *Suffix
	Sibling     *Suffix
	Suffixes    *Suffix
	baseDecl    *VarDecl
}

// NewVarDecl creates and initialize a new variable type declaration.
func NewVarDecl(name string, typ pmmp.ValueType) *VarDecl {
	decl := &VarDecl{}
	decl.suffixName = name
	decl.Tag = *runtime.NewTag(name)
	decl.Tag.Typ = int8(typ)
	decl.Tag.UData = decl
	decl.baseDecl = decl // this pointer should never be nil
	T().P("decl", decl.FullName()).Debugf("atomic variable type declaration created")
	return decl
}

// AsSuffix returns a variable declaration as a Suffix. Every variable
// declaration is also a suffix, although one without a parent.
func (d *VarDecl) AsSuffix() *Suffix {
	return &d.Suffix
}

// AsTag returns the variable declaration as a tag (for symbol tables).
func (d *VarDecl) AsTag() *runtime.Tag {
	return &d.Tag
}

// DeclFromTag returns the declaration represented by a tag.
func DeclFromTag(tag *runtime.Tag) *VarDecl {
	return tag.UData.(*VarDecl)
}

func (d *VarDecl) String() string {
	return fmt.Sprintf("<decl %s/%s>", d.FullName(), d.Type().String())
}

// IsBaseDecl is a prediacte: is this suffix the base tag?
func (s *Suffix) IsBaseDecl() bool {
	return s.Parent == nil
}

// Type returns the variable's type.
func (s *Suffix) Type() pmmp.ValueType {
	return pmmp.ValueType(s.baseDecl.Tag.Typ)
}

// BaseDecl returns the base type declaration for a tag.
func (s *Suffix) BaseDecl() *VarDecl {
	return s.baseDecl
}

// SuffixName gets the isolated name of declaration partial
// (tag, array or suffix).
func (s *Suffix) SuffixName() string {
	if s.isSubscript {
		return "[]"
	} else if s.Parent != nil {
		return "." + s.suffixName
	}
	return s.BaseDecl().Name()
}

// FullName gets
// the full name of a type declaration, starting with the base tag.
// x <- array <- suffix(a)  gives "x[].a", which is a bit more verbose
// than MetaFont's response. I prefer this one.
//
func (s *Suffix) FullName() string {
	if s.Parent == nil {
		return s.baseDecl.Tag.Name()
	} // we are in a declaration for a complex type
	var str bytes.Buffer
	str.WriteString(s.Parent.FullName()) // recursive
	if s.isSubscript {
		str.WriteString("[]")
	} else {
		str.WriteString(".") // MetaFont suppresses this if following a subscript partial
		str.WriteString(s.suffixName)
	}
	return str.String()
}

// GetBaseType gets the type of the base tag.
// func (s *Suffix) GetBaseType() pmmp.ValueType {
// 	return s.baseDecl.Type()
// }

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
func CreateSuffix(name string, typ pmmp.ValueType, parent *Suffix) *Suffix {
	if typ != pmmp.SubscriptType && typ != pmmp.SuffixType {
		panic("Suffix must either be of type SuffixType or SubscriptType")
	}
	if parent != nil { // check if already exists as child of parent
		if parent.Suffixes != nil {
			ch := parent.Suffixes
			for ch != nil { // check all direct suffixes
				if (typ == pmmp.SuffixType && ch.suffixName == name) ||
					(ch.isSubscript && typ == pmmp.SubscriptType) {
					T().P("decl", ch.FullName()).Debugf("variable type already declared")
					return ch // we're done
				}
				ch = ch.Sibling
			}
		}
	}
	s := &Suffix{ // not found => create a new one
		suffixName:  name,
		isSubscript: typ == pmmp.SubscriptType,
	}
	if parent == nil { // then wrap into a var decl
		if s.isSubscript { // can't have a subscript as a top-level var decl
			panic("Trying to have standalone subscript")
		}
		d := NewVarDecl(name, pmmp.NumericType) // numeric is default in MetaPost
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
	b.WriteString(fmt.Sprintf("%s : %s\n", s.FullName(), s.baseDecl.Type().String()))
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
	runtime.Tag            // store by normalized name
	id          int32      // unique ID
	decl        *Suffix    // type declaration for this variable
	subscripts  []float64  // list of subscript value, first to last
	Value       pmmp.Value // value of type numeric, pair, path, …
}

// CreateVarRef creates a variable reference. Low level method.
func CreateVarRef(decl *Suffix, value pmmp.Value, indices []float64) *VarRef {
	T().Debugf("creating %s var for %v", decl.Type().String(), decl.FullName())
	v := &VarRef{
		decl:       decl,
		subscripts: indices,
		Value:      value,
	}
	v.Tag = *runtime.NewTag(v.FullName())
	v.Typ = int8(decl.Type())
	v.id = serialCounter.Get()
	//T().Debugf("created var ref: subscripts = %v", indices)
	v.Tag.UData = v // link back to var from tag
	if decl.Type() == pmmp.PairType {
		return createPairTypeVarRef(v, decl, value, indices)
	}
	return v
}

func (v *VarRef) String() string {
	return fmt.Sprintf("<var %s=%v w/ %s>", v.FullName(), v.Value, v.decl.FullName())
}

// Name returns the full nomalized name, i.e. "x[2].r".
// This enables us to store the variable in a symbol table.
// Overrides Name of interface Symbol.
//
// func (v *VarRef) Name() string {
// if len(v.cachedName) == 0 {
// 	v.cachedName = v.FullName()
// }
// return v.cachedName
// }

// ID gets the variable's ID.
func (v *VarRef) ID() int32 {
	return v.id
}

// ResetID sets the variables's ID to a new and unused value.
func (v *VarRef) ResetID() int32 {
	serial := serialCounter.Get()
	v.id = serial
	return v.id
}

// Type returns the variable's type.
func (v *VarRef) Type() pmmp.ValueType {
	if v.decl != nil {
		return v.decl.Type()
	}
	return pmmp.Undefined
}

// Declaration returns the variable reference's base tag declaration.
func (v *VarRef) Declaration() *VarDecl {
	return v.decl.baseDecl
}

// AsTag returns the variable reference as a tag to store it into a symbol table.
func (v *VarRef) AsTag() *runtime.Tag {
	return &v.Tag
}

// VarFromTag returns a tag as a variable reference.
// If the tag stands for a pair partial, VarFromTag will return a reference
// to the parent (pair) variable.
func VarFromTag(tag *runtime.Tag) *VarRef {
	return tag.UData.(*VarRef)
}

// SuffixesString returns a variable's name without the leading base tag name.
// Strip the base tag string off of a variable and return all the suffxies
// as string.
//
func (v *VarRef) SuffixesString() string {
	declbase := v.decl.baseDecl
	basename := declbase.Name()
	fullstring := v.FullName()
	return fullstring[len(basename):]
}

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
// type PairPartRef struct {
// 	runtime.Tag
// 	id      int32      // serial ID
// 	Pairvar *VarRef    // pair parent
// 	Value   pmmp.Value // value of type numeric
// }
//
type pairVarValues struct { // has to satisfy interface pmmp.Value
	values [2]PairPartValue // but is not used as an actual value
}

func (pv *pairVarValues) Self() pmmp.ValueBase {
	return pmmp.ValueBase{V: pv}
}

func (pv *pairVarValues) IsKnown() bool {
	return pv.values[0].Value.IsKnown() && pv.values[1].Value.IsKnown()
}

func (pv *pairVarValues) Type() pmmp.ValueType {
	return pmmp.Undefined
}

func (pv *pairVarValues) xPart() *PairPartValue {
	return &pv.values[0]
}

func (pv *pairVarValues) yPart() *PairPartValue {
	return &pv.values[1]
}

// PairPartValue is a pseudo-variable to hold a pair part value of a
// pair variable. It is wrapped into a variable struct to be able to feed
// it into the LEQ solver.
type PairPartValue struct {
	runtime.Tag
	id       int32
	variable *VarRef
	Value    pmmp.Value
}

func newPairVarValues(v *VarRef) *pairVarValues {
	ppv := &pairVarValues{}
	ppv.values[0].id = v.id
	ppv.values[0].variable = v
	ppv.values[0].Tag.UData = ppv.values[0]
	ppv.values[1].id = serialCounter.Get()
	ppv.values[1].variable = v
	ppv.values[1].Tag.UData = ppv.values[1]
	return ppv
}

// CreatePairTypeVarRef creates a pair variable reference. Low level method.
func createPairTypeVarRef(v *VarRef, decl *Suffix, value pmmp.Value, indices []float64) *VarRef {
	T().Debugf("extending pair var for %v", decl.FullName())
	// v := &VarRef{
	// 	decl:       decl,
	// 	subscripts: indices,
	// 	Value:      value,
	// }
	// v.Tag.Name = decl.name
	v.Typ = int8(pmmp.PairType)
	//
	T().Debugf("creating pair values proxy")
	pv := newPairVarValues(v)
	v.Value = pv
	//
	v.Set(value)
	// if pair, ok = value.(arithm.Pair); ok {
	// 	T().Debugf("setting value of pair var to %v", pair)
	// 	xpart.Value = pair.X()
	// 	ypart.Value = pair.Y()
	// } else {
	// 	xpart.Value = math.NaN()
	// 	ypart.Value = math.NaN()
	// }
	return v
}

// Name returns a string for a pair part.
// Pair parts (x-part or y-part) return the name of their parent pair symbol,
// prepending "xpart" or "ypart" respectively. This name is constant and
// may be used to store the pair part in a symbol table.
//
func (ppv *PairPartValue) Name() string {
	if ppv.variable.id == ppv.id {
		return "xpart " + ppv.variable.FullName()
	}
	return "ypart " + ppv.variable.FullName()
}

// Type returns the type of a pair part, which is always numeric.
func (ppv *PairPartValue) Type() pmmp.ValueType {
	return pmmp.NumericType
}

// AsTag returns a pair part as a tag.
func (ppv *PairPartValue) AsTag() *runtime.Tag {
	return &ppv.Tag
}

// PPVFromTag returns the pair partial from a tag.
func PPVFromTag(tag *runtime.Tag) *PairPartValue {
	return tag.UData.(*PairPartValue)
}

func (ppv *PairPartValue) String() string {
	return "<" + ppv.Name() + fmt.Sprintf("=%v>", ppv.Value)
}

// IsPair is a predicate: is this variable of type pair?
func (v *VarRef) IsPair() bool {
	return v.Type() == pmmp.PairType
}

// XPart gets the x-part of a pair variable
func (v *VarRef) XPart() *PairPartValue {
	if !v.IsPair() {
		T().P("var", v.Name).Errorf("cannot access x-part of non-pair")
		return nil
	}
	if v.Value == nil {
		panic("pair variable must never be without values proxy")
	}
	values := v.Value.(*pairVarValues)
	return values.yPart()
}

// YPart gets the y-part of a pair variable
func (v *VarRef) YPart() *PairPartValue {
	if !v.IsPair() {
		T().P("var", v.Name).Errorf("cannot access x-part of non-pair")
		return nil
	}
	if v.Value == nil {
		panic("pair variable must never be without values proxy")
	}
	values := v.Value.(*pairVarValues)
	return values.yPart()
}

// func (v *VarRef) YPart() *PairPartRef {
// 	if !v.IsPair() {
// 		T().P("var", v.Name).Errorf("cannot access y-part of non-pair")
// 		return nil
// 	}
// 	return PPFromTag(v.Children)
// }

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
// func (ppart *PairPartRef) IsKnown() bool {
// 	return (ppart.Value != nil)
// }

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
			suffixes = append(suffixes, sfx.suffixName)
		}
	}
	// now reverse the slice of suffixes
	for i := 0; i < len(suffixes)/2; i++ { // swap suffixes in place
		j := len(suffixes) - i - 1
		suffixes[i], suffixes[j] = suffixes[j], suffixes[i]
	}
	return strings.Join(suffixes, "")
}

// --- Variable ref methods --------------------------------------------------

// HasKnownValue is a predicate: has this variable a known value?
func (v *VarRef) HasKnownValue() bool {
	if !v.IsPair() {
		return v.Value.IsKnown()
	}
	return v.Value.(*pairVarValues).IsKnown()
}

// Get gets a variable's value.
func (v *VarRef) Get() pmmp.Value {
	if !v.IsPair() {
		return v.Value
	}
	if v.Value == nil {
		return pmmp.NullPair()
	}
	return pmmp.NewPair(v.XPart().Value.Self().AsNumeric(),
		v.YPart().Value.Self().AsNumeric())
}

// Set sets a variable's value.
func (v *VarRef) Set(val pmmp.Value) {
	T().P("var", v.Name).Debugf("new value: %v", val)
	if !v.IsPair() {
		switch v.Type() {
		case pmmp.NumericType:
			if val == nil {
				v.Value = pmmp.Numeric{}
			} else {
				v.Value = val
			}
		}
		return
	}
	T().Debugf("setting pair values")
	var xpart *PairPartValue = v.XPart()
	var ypart *PairPartValue = v.YPart()
	if val == nil {
		xpart.Value = pmmp.Numeric{}
		ypart.Value = pmmp.Numeric{}
	} else {
		if pairval, ok := val.(pmmp.Pair); ok {
			xpart.Value = pairval.XNumeric()
			ypart.Value = pairval.YNumeric()
		}
	}
}

// PullValue pulls a variable's value up to a containig pair variable.
// Whenever a pair part (x-part or y-part) is set, it sends a message to
// the parent pair variable to pull the value. If both parts are known and
// a numeric value is set, the parent pair creates a combined pair value.
//
// func (v *VarRef) PullValue() {
// 	if v.IsPair() {
// 		var ppart1, ppart2 *PairPartRef
// 		ppart1 = PPFromTag(v.Sibling)
// 		ppart2 = PPFromTag(v.Children)
// 		if ppart1 != nil && ppart2 != nil {
// 			if ppart1.Value != nil && ppart2.Value != nil {
// 				v.Value = arithm.P(ppart1.Value.(float64), ppart2.Value.(float64))
// 				T().P("var", v.Name).Debugf("pair value = %s",
// 					v.Value.(arithm.Pair).String())
// 			}
// 		}
// 	}
// }

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
			xvalue = fmt.Sprintf("%v", xpart)
		}
		ypart := v.YPart().Value
		if ypart == nil {
			yvalue = fmt.Sprintf("ypart %s", v.Name())
		} else {
			yvalue = fmt.Sprintf("%v", ypart)
		}
		return fmt.Sprintf("(%s,%s)", xvalue, yvalue)
	}
	if v.HasKnownValue() {
		return fmt.Sprintf("%v", v)
	}
	return "<numeric>"
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
	oldserial := v.id
	v.id = serialCounter.Get()
	if v.IsPair() {
		ypartid := serialCounter.Get()
		v.XPart().id = v.id
		v.YPart().id = ypartid
	}
	return oldserial
}

// --- Unique ID for tags ----------------------------------------------------

// UniqueID is a counter type.
type UniqueID struct {
	counter int32
}

// Get fetches a new unique id from this counter.
func (c *UniqueID) Get() int32 {
	for {
		val := atomic.LoadInt32(&c.counter)
		if atomic.CompareAndSwapInt32(&c.counter, val, val+1) {
			return val
		}
	}
}

var serialCounter UniqueID // global serial counter
