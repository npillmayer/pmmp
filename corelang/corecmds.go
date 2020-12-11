package corelang

// Internal commands for our core DSL, borrowing from MetaFont/MetaPost.

import (
	"bytes"
	"fmt"
	"math"
	"strings"

	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/npillmayer/arithm"
	"github.com/npillmayer/arithm/polyn"
	"github.com/npillmayer/gorgo/runtime"
	"github.com/npillmayer/pmmp/variables"
)

// WhateverDeclaration is the global declaration of 'whatever', used to
// instantiate anonymous whatever-variables.
var WhateverDeclaration *variables.VarDecl

// === Handling Variables and Constants ======================================

/*
CollectVarRefParts constructs
a valid variable reference string from parts on the stack.

Collect fragments of a variable reference, e.g. "x[k+1]r".
Subscripts should be found on the expression stack and inserted as
numeric constants, i.e. resulting in "x[5]r" (if k=4).

Parameter t is the text of the variable ref literal, e.g. "x[k+1]r".
It is split by the parser into:

  . "x" -> TAG x
  . subscript { k+1 }
  . "r" -> TAG r

*/
func CollectVarRefParts(rt *runtime.Runtime, t string, children []antlr.Tree) string {
	var vname bytes.Buffer
	for _, ch := range children {
		T().Debugf("collecting var ref part: %s", getCtxText(ch))
		if isTerminal(ch) { // just copy string parts to output
			T().Debugf("adding suffix verbatim: %s", getCtxText(ch))
			vname.WriteString(ch.(antlr.ParseTree).GetText())
		} else { // non-terminal is a subscript-expression
			subscript, ok := exprStack(rt).Pop() // take subscript from stack
			if !ok {
				T().P("var", t).Errorf("expected subscript on expression stack")
				T().P("var", t).Errorf("substituting 0 instead")
				vname.WriteString("[0]")
			} else {
				c, isconst := subscript.XPolyn.IsConstant()
				if !isconst { // we cannot handle unknown subscripts
					T().P("var", t).Errorf("subscript must be known numeric")
					T().P("var", t).Errorf("substituting 0 for %s",
						exprStack(rt).TraceString(subscript))
					vname.WriteString("[0]")
				} else {
					vname.WriteString("[")
					vname.WriteString(fmt.Sprintf("%g", c))
					vname.WriteString("]")
				}
			}
		}
	}
	varname := vname.String()
	T().P("var", varname).Debugf("collected parts")
	return varname
}

/*
MakeCanonicalAndResolve gets or creates a variable reference.
To get the canonical representation of
the variable reference, we parse it and construct a small AST. This AST
is fed into GetVarRefFromVarSyntax(). The resulting variable reference
struct is used to find the memory location of the variable reference.

Example:

	vref := MakeCanonicalAndResolve(rt, "a2r", true)
	// now vref.String() gives something like:
	//      "<var a[2].r=<nil> w/ <decl a[].r/numeric>>"

If a variable has been undeclared and is now created, the top-most scope
and memory-frame will hold the newly created variable.
*/
func MakeCanonicalAndResolve(rt *runtime.Runtime, v string, create bool) (
	*variables.VarRef, error) {
	//
	panic("Not yet implemented")
}

// 	var vref *variables.VarRef
// 	resolve, err := varparse.ParseVariableName(v)
// 	if err != nil {
// 		return nil, err
// 	}
// 	vref = resolve.VariableReference(rt.ScopeTree)
// 	if vref != nil {
// 		vref, _ = FindVariableReferenceInMemory(rt, vref, create)
// 	}
// 	return vref, nil
// }

// AllocateVariableInMemory allocates
// a variable in a given memory frame. Existing variable references in
// this memory frame will be overwritten !
// Clients should probably first call FindVariableReferenceInMemory(vref).
func AllocateVariableInMemory(vref *variables.VarRef,
	mf *runtime.DynamicMemoryFrame) *variables.VarRef {
	//
	mf.SymbolTable.InsertTag(vref.AsTag())
	T().P("var", vref.FullName()).Debugf("allocating variable in %s", mf.Name)
	return vref
}

/*
FindVariableReferenceInMemory searches for a current variable reference.
Given a variable reference, locate an incarnation in a memory frame. The
frame is determined by the variable's declaring scope: search for the top
frame linked to the scope.

Variable references live in memory frames. Memory frames correspond to
scopes. To find a variable reference -- i.e. a living variable with a possible
value -- we have to proceed as follows:

(1) find the variable declaration in a scope, beginning at the top

(2) find the most recent memory frame pointing to this scope

(3) find a variable reference with the correct name in the memory frame

(4) if no reference/incarnation exists, create one

Parameter doAlloc: should step (4) be performed ?
*/
func FindVariableReferenceInMemory(rt *runtime.Runtime, vref *variables.VarRef, doAlloc bool) (
	*variables.VarRef, *runtime.DynamicMemoryFrame) {
	//
	if vref.Declaration() == nil {
		T().P("var", vref.FullName()).Errorf("attempt to store variable without decl. in memory")
		return vref, nil
	}
	var sym *variables.VarRef
	var memframe *runtime.DynamicMemoryFrame
	tagname := vref.Declaration().FullName()
	tag, scope := rt.ScopeTree.Current().ResolveTag(tagname)
	if tag != nil { // found tag declaration in scope
		memframe = rt.MemFrameStack.FindMemoryFrameForScope(scope)
		varname := vref.Name()
		T().P("var", varname).Debugf("var in ? %s", memframe)
		s := memframe.SymbolTable.ResolveTag(varname)
		if s == nil { // no variable ref incarnation => create one
			T().P("var", varname).Debugf("not found in memory")
			if doAlloc {
				sym = AllocateVariableInMemory(vref, memframe)
			}
		} else { // already present, return this one
			T().P("var", varname).Debugf("variable already present in memory")
			sym = variables.VarFromTag(s)
		}
	} else {
		// this should never happen: we could neither find nor construct a var decl
		panic(fmt.Sprintf("declaration for %s mysteriously vanished...", tagname))
	}
	return sym, memframe
}

// PushVariable puts a variable onto the expression stack.
// The expression stack knows nothing about the interpreter's symbols, except
// the few properties of interface Symbol. The expression stack deals with
// polynomials and serial IDs of variables.
//
// Push a variable (numeric or pair type) onto the expression stack.
//
func PushVariable(rt *runtime.Runtime, vref *variables.VarRef, asLValue bool) {
	if vref.IsPair() {
		if vref.IsKnown() && !asLValue {
			PushConstant(rt, vref) // put constant on expression stack
		} else {
			xpart, ypart := vref.XPart(), vref.YPart()
			exprStack(rt).PushVariable(xpart.AsTag(), ypart.AsTag())
		}
	} else {
		if vref.IsKnown() && !asLValue {
			PushConstant(rt, vref) // put constant on expression stack
		} else {
			exprStack(rt).PushVariable(vref.AsTag(), nil)
		}
	}
}

// PushConstant pushes a constant (numeric or pair type) onto the expression stack.
func PushConstant(rt *runtime.Runtime, vref *variables.VarRef) {
	switch vref.Type() {
	case variables.NumericType:
		exprStack(rt).PushConstant(vref.Value.(float64))
	case variables.PairType:
		x := vref.XPart().Value
		y := vref.YPart().Value
		pair := arithm.P(x.(float64), y.(float64))
		exprStack(rt).PushPairConstant(pair)
	case variables.PathType:
		exprStack(rt).PushOtherConstant(vref.Value)
	}
}

// GetVariableFromExpression converts a variable to its expression value.
// The expression stack knows nothing about the interpreter's symbols, except
// the few properties of interface Symbol. The expression stack deals with
// polynomials and serial IDs of variables. To get back from IDs to
// variable references, we ask the expression stack for a Symbol (from an
// ID). If the variable is of type pair, the Symbol may be a pair part (x-part
// or y-part). Parts point to their parent symbol, thus giving us the
// variable reference.
func GetVariableFromExpression(rt *runtime.Runtime, e *ExprNode) *variables.VarRef {
	var v *variables.VarRef
	if sym := exprStack(rt).GetVariable(e); sym != nil {
		// var part *variables.PairPartRef
		// var ok bool
		// if part, ok = sym.(*variables.PairPartRef); ok {
		// 	sym = part.Pairvar
		// }
		v = variables.VarFromTag(sym)
		T().P("var", v.Name()).Debugf("variable of type %s", v.Type().String())
	}
	return v
}

// EncapsulateVariable makes a variable an anonymous value.
// A variable which goes out of scope becomes a "capsule". We send a message
// to the expression stack to forget the Symbol(s) for the ID(s) of a
// variable. Variables are of type numeric or pair.
func EncapsulateVariable(rt *runtime.Runtime, v *variables.VarRef) {
	exprStack(rt).EncapsuleVariable(v.ID)
	if v.IsPair() {
		var ypart *variables.PairPartRef = variables.PPFromTag(v.Children)
		exprStack(rt).EncapsuleVariable(ypart.ID)
	}
}

// EncapsulateVarsInMemory makes all variables in a memory frame "capsules".
//
// When a memory frame is popped from the stack, the local variables living
// in the frame have to be made "capsules". This is necessary, because they
// may still be relevant to the LEQ-solver. The LEQ will finally decide
// when to abondon the "zombie" variable.
func EncapsulateVarsInMemory(rt *runtime.Runtime, mf *runtime.DynamicMemoryFrame) {
	mf.SymbolTable.Each(func(name string, sym *runtime.Tag) {
		vref := variables.VarFromTag(sym)
		T().P("var", vref.FullName()).Debugf("encapsule")
		exprStack(rt).EncapsuleVariable(vref.ID) // vref is now capsule
	})
}

// LoadBuiltinSymbols loads a bunch of pre-defined symbols
// into a scope (usually the global scope).
// Additionally loads initial Lua definitions.
func LoadBuiltinSymbols(rt *runtime.Runtime, scripting *Scripting) {
	originDef := Declare(rt, "origin", variables.PairType)
	origin := arithm.P(0, 0)
	_ = Variable(rt, originDef, origin, nil, true)
	upDef := Declare(rt, "up", variables.PairType)
	up := arithm.P(0, 1)
	_ = Variable(rt, upDef, up, nil, true)
	downDef := Declare(rt, "down", variables.PairType)
	down := arithm.P(0, -1)
	_ = Variable(rt, downDef, down, nil, true)
	rightDef := Declare(rt, "right", variables.PairType)
	right := arithm.P(1, 0)
	_ = Variable(rt, rightDef, right, nil, true)
	leftDef := Declare(rt, "left", variables.PairType)
	left := arithm.P(-1, 0)
	_ = Variable(rt, leftDef, left, nil, true)
	_ = Declare(rt, "p", variables.PairType)
	_ = Declare(rt, "q", variables.PairType)
	_ = Declare(rt, "P", variables.PathType)
	w := Declare(rt, "_whtvr", variables.NumericType) // 'whatever' variables
	// make whatever[]
	WhateverDeclaration = variables.CreateSuffix("<array>", variables.SubscriptType,
		w.AsSuffix()).BaseDecl()
}

// === Commands ==============================================================

/*
Assign is a variable assignment.

   assignment : lvalue ASSIGN numtertiary

(1) Retract lvalue from the resolver's table (make a capsule)

(3) Unset the value of lvalue

(3) Re-incarnate lvalue (get a new ID for it)

(4) If type is numeric or pair: Create equation on expression stack,
else assign a path value to a path variable.
*/
func Assign(rt *runtime.Runtime, lvalue *variables.VarRef, e *ExprNode) {
	varname := lvalue.Name()
	oldserial := lvalue.ID
	T().P("var", varname).Debugf("assignment of lvalue #%d", oldserial)
	EncapsulateVariable(rt, lvalue)
	vref, mf := FindVariableReferenceInMemory(rt, lvalue, false)
	vref.SetValue(nil) // now lvalue is unset / unsolved
	T().P("var", varname).Debugf("unset in %v", mf)
	vref.Reincarnate()
	T().P("var", vref.Name()).Debugf("new lvalue incarnation #%d", vref.ID)
	if vref.Type() == variables.PathType {
		vref.SetValue(e.Other)
	} else { // create linear equation
		PushVariable(rt, vref, false) // push LHS on stack
		exprStack(rt).Push(e)         // push RHS on stack
		exprStack(rt).EquateTOS2OS()  // construct equation
	}
}

// Save a tag within a group. The tag will be restored at the end of the
// group. Save-commands within global scope will be ignored.
// This method simply creates a var decl for the tag in the current scope.
func Save(rt *runtime.Runtime, tag string) {
	sym, scope := rt.ScopeTree.Current().ResolveTag(tag)
	if sym != nil { // already found in scope stack
		T().P("tag", tag).Debugf("save: found tag in scope %s",
			scope.Name)
	}
	T().Debugf("declaring %s in current scope", tag)
	rt.ScopeTree.Current().DefineTag(tag)
}

// Declare a tag to be of type tp.
//
// If the tag is not declared, insert a new symbol in global scope. If a
// declaration already exists, erase all variables and re-enter a declaration
// (MetaFont semantics). If the tag has been "saved" in the current or in an outer
// scope, make this tag a new undefined symbol.
//
func Declare(rt *runtime.Runtime, tag string, tp variables.VariableType) *variables.VarDecl {
	sym, scope := rt.ScopeTree.Current().ResolveTag(tag)
	if sym != nil { // already found in scope stack
		T().P("tag", tag).Debugf("declare: found tag in scope %s", scope.Name)
		T().P("decl", tag).Debugf("variable already declared - re-declaring")
		// Erase all existing variables and re-define symbol
		sym, _ = scope.DefineTag(tag)
		//variables.DeclFromTag(sym).ChangeType(tp)
		sym.ChangeType(int8(tp))
	} else { // enter new symbol in global scope
		scope = rt.ScopeTree.Globals()
		sym, _ = scope.DefineTag(tag)
		sym.ChangeType(int8(tp))
	}
	T().P("decl", sym.Name).Debugf("declared symbol in %s", scope.Name)
	return variables.DeclFromTag(sym)
}

// Variable creates a
// variable reference. Parameters are the declaration for the variable,
// a value and a flag, indicating if this variable should go to global memory.
// The subscripts parameter is a slice of array-subscripts, if the variable
// declaration is of array (complex) type.
func Variable(rt *runtime.Runtime, decl *variables.VarDecl, value interface{},
	subscripts []float64, global bool) *variables.VarRef {
	//
	var v *variables.VarRef
	if decl.Type() == variables.NumericType {
		v = variables.CreateVarRef(decl.AsSuffix(), value, subscripts)
	} else {
		v = variables.CreatePairTypeVarRef(decl.AsSuffix(), value, subscripts)
	}
	if global {
		rt.MemFrameStack.Globals().SymbolTable.InsertTag(v.AsTag())
	} else {
		rt.MemFrameStack.Current().SymbolTable.InsertTag(v.AsTag())
	}
	return v
}

// Whatever creates
// a whatever anonymous variable. In MetaFont this is a macro, but
// it is a frequent use case, so we put it in the core.
func Whatever(rt *runtime.Runtime) *variables.VarRef {
	var vref *variables.VarRef
	sym, _ := rt.ScopeTree.Globals().ResolveTag("_whtvr")
	if sym == nil {
		T().Errorf("'whatever'-variable not correctly initialized")
	} else {
		inx := []float64{1.0}
		whateverCounter++
		inx[0] = float64(whateverCounter)
		vref = variables.CreateVarRef(WhateverDeclaration.AsSuffix(), nil, inx)
	}
	return vref
}

// Counter for 'whatever' anonymous variables.
var whateverCounter int64

// CallFunc applies a
// (math or scripting) function, given by name, to a known/constant argument.
// Internal math functions are floor(), ceil() and sqrt(). Other function names
// will be delegated to the scripting subsystem (Lua).
//
// Lua functions will return just one value (of type numeric, pair or path).
//
func CallFunc(val interface{}, fun string, scripting *Scripting) (*ExprNode, []*variables.VarRef) {
	n := 0.0
	if strings.HasPrefix(fun, "@") {
		fun = strings.TrimLeft(fun, "@")
		T().P("func", fun).Debugf("calling Lua scripting subsytem")
		r, err := scripting.CallHook(fun, val)
		if err == nil {
			it := r.Iterator() // iterator over return values
			if it.Next() {     // go to first return value
				e, vars := it.ValueAsExprNode() // unpack first return value only
				return e, vars
			}
		} else {
			T().P("func", fun).Errorf("scripting error: %v", err.Error())
		}
	} else {
		switch fun {
		case "floor":
			n = val.(float64)
			n = math.Floor(n)
		case "ceil":
			n = val.(float64)
			n = math.Ceil(n)
		case "sqrt":
			T().P("func", fun).Errorf("function not yet implemented")
		default:
			T().P("func", fun).Errorf("function not implemented")
		}
	}
	p := polyn.NewConstantPolynomial(n)
	e := NewNumericExpression(p)
	return e, nil
}

// CallVardef calls the
// Lua script for a vardef variable. The parameters of the call will
// be the suffixes of the variable, i.e. all subscripts and tags after the base
// tag.
//
// Return values are wrapped into an expression node. If variables are part of
// the returned expressions, they are packed into an array of variable
// references.
//
func CallVardef(vref *variables.VarRef, scripting *Scripting) (*ExprNode, []*variables.VarRef) {
	basedecl := vref.Declaration()
	suffixes := vref.SuffixesString()
	T().Debugf("vardef call %s(%s)", basedecl.FullName(), suffixes)
	r, err := scripting.Call("vardef", basedecl.FullName(), suffixes)
	if err == nil {
		it := r.Iterator() // iterator over return values
		if it.Next() {     // go to first return value
			e, vars := it.ValueAsExprNode() // unpack first return value only
			return e, vars
		}
	} else {
		T().P("vardef", basedecl.FullName()).Errorf("scripting error: %v", err.Error())
	}
	return nil, nil
}

// Begingroup is the
// MetaPost begingroup command: push a new scope and memory frame.
// Clients may supply a name for the group, otherwise it will be set
// to "group".
func Begingroup(rt *runtime.Runtime, name string) (*runtime.Scope, *runtime.DynamicMemoryFrame) {
	if name == "" {
		name = "group"
	}
	groupscope := rt.ScopeTree.PushNewScope(name) // , variables.NewVarDecl)
	groupmf := rt.MemFrameStack.PushNewMemoryFrame(name, groupscope)
	return groupscope, groupmf
}

// Endgroup is the
// MetaPost endgroup command: pop scope and memory frame of group.
func Endgroup(rt *runtime.Runtime) {
	mf := PopScopeAndMemory(rt)
	EncapsulateVarsInMemory(rt, mf)
}

// PopScopeAndMemory decreases the grouping level.
// We pop the topmost scope and topmost memory frame. This happens after
// a group is left.
//
// Returns the previously topmost memory frame.
//
func PopScopeAndMemory(rt *runtime.Runtime) *runtime.DynamicMemoryFrame {
	hidden := rt.ScopeTree.PopScope()
	hidden.Name = "(hidden)"
	mf := rt.MemFrameStack.PopMemoryFrame()
	if mf.Scope != hidden {
		T().P("mem", mf.Name).Errorf("groups out of sync?")
	}
	return mf
}

// === Show Commands =========================================================

// Showvariable shows all declarations and references for a tag.
func Showvariable(rt *runtime.Runtime, tag string) string {
	sym, scope := rt.ScopeTree.Current().ResolveTag(tag)
	if sym == nil {
		return fmt.Sprintf("%s : tag\n", tag)
	}
	v := variables.DeclFromTag(sym)
	var b *bytes.Buffer
	b = v.ShowDeclarations(b)
	if mf := rt.MemFrameStack.FindMemoryFrameForScope(scope); mf != nil {
		for _, v := range mf.SymbolTable.Table {
			vref := variables.VarFromTag(v)
			if vref.Declaration().Tag() == sym {
				s := fmt.Sprintf("%s = %s\n", vref.FullName(), vref.ValueString())
				b.WriteString(s)
			}
		}
	}
	return b.String()
}

// === Utilities =============================================================

// Unit2numeric converts a unit of length (cm, mm, pt, in, …)
// to internal scaled points.
//
// TODO complete this.
func Unit2numeric(u string) float64 {
	switch u {
	case "in":
		return 0.01388888
	}
	return 1.0
}

// ScaleDimension scales a numeric value by a unit.
func ScaleDimension(dimen float64, unit string) float64 {
	u := Unit2numeric(unit)
	return dimen * u
}

// --- Utilities -------------------------------------------------------------

func isTerminal(x interface{}) bool {
	panic("TODO implement isTerminal")
}

func getCtxText(ctx interface{}) string {
	return "getCtxText-TODO"
}

func exprStack(rt *runtime.Runtime) *ExprStack {
	return rt.UData.(*ExprStack)
}
