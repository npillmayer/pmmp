/*
Package varparse implements functions to create variable declarations
and references from syntax trees.



BSD License

Copyright (c) 2017-21, Norbert Pillmayer

All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions
are met:

1. Redistributions of source code must retain the above copyright
notice, this list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright
notice, this list of conditions and the following disclaimer in the
documentation and/or other materials provided with the distribution.

3. Neither the name of the software nor the names of its contributors
may be used to endorse or promote products derived from this software
without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE. */
package varparse

import (
	"github.com/npillmayer/gorgo/terex"
	"github.com/npillmayer/pmmp/variables"
	"github.com/npillmayer/schuko/gtrace"
	"github.com/npillmayer/schuko/tracing"
)

// T traces to the global syntax tracer
func T() tracing.Trace {
	return gtrace.SyntaxTracer
}

// === Variable Parser =======================================================

// After having created an abstract syntax tree (AST) for a variable identifier,
// we will construct a variable reference from the parse tree. This var ref
// has an initial serial which is unique. This may not be what you want:
// usually you will try to find an existing incarnation (with a lower serial)
// in the memory (see method FindVariableReferenceInMemory).
//
// We walk the AST using a listener (varParseListener).
//
func getVarRefFromVarSyntax(el terex.Element) *variables.VarRef {
	//
	el.Dump(tracing.LevelInfo)
	return nil
}

type suffixBuilder func(node terex.Atom, parent terex.Atom) terex.Element

func createSuffix(node terex.Atom, parent terex.Atom) terex.Element {
	return terex.Elem(nil)
}

// Traverser in pre-order.
func traverseAST(e terex.Element, parent terex.Atom, sb suffixBuilder) terex.Element {
	if e.IsAtom() {
		return sb(e.AsAtom(), parent)
	}
	list := e.AsList()
	curNode := list.First()
	sb(curNode, parent)
	//
	rest := list.Rest()
	for rest != nil {
		node := rest.Car
		if node.Type() != terex.ConsType {
			//
			sb(node, curNode)
		}
		rest = rest.Cdr
	}
	return terex.Elem(e)
}

/*
Listener callback, receiving a complete variable reference.

   variable : tag (suffix | subscript)* MARKER

A variable has been referenced. We will have to find the declaration of
this variable and push a variable reference onto the expression stack.
A variable reference looks something like this: "x2a.b" or "y[3.14]".

Some complications may arise:

- No declaration for the variable's tag can be found: we will create
a declaration for the tag in global scope with type numeric.

- The declaration is incomplete, i.e. the tag is declared, but not the
suffix(es). We will extend the declaration appropriately.

- We will have to create a subscript vector for the var reference.
we'll collect them (together with suffixes) in a list.

Example:
Parser reads "x2a", thus

   tag="x" + subscript="2" + suffix="a"

We create (if not yet known)

   vl.def="x[].a"

and

   vl.ref="x[2].a"

The MARKER will be ignored.
*/
/*
func (vl *varParseListener) ExitVariable(ctx *grammar.VariableContext) {
	tag := ctx.Tag().GetText()
	T().P("tag", tag).Debugf("looking for declaration for tag")
	sym, scope := vl.scopeTree.Current().ResolveSymbol(tag)
	if sym != nil {
		vl.def = sym.(*variables.PMMPVarDecl) // scopes are assumed to create these
		T().P("decl", vl.def.GetFullName()).Debugf("found %v in scope %s", vl.def, scope.GetName())
	} else { // variable declaration for tag not found => create it
		sym, _ = vl.scopeTree.Globals().DefineSymbol(tag)
		vl.def = sym.(*variables.PMMPVarDecl)      // scopes are assumed to create these
		vl.def.SetType(int(variables.NumericType)) // un-declared variables default to type numeric
		T().P("decl", vl.def.GetName()).Debugf("created %v in global scope", vl.def)
	} // now def declaration of <tag> is in vl.def
	// produce declarations for suffixes, if necessary
	it := vl.suffixes.Iterator()
	subscrCount := 0
	for it.Next() {
		i, vs := it.Index(), it.Value().(varsuffix)
		T().P("decl", vl.def.GetFullName()).Debugf("appending suffix #%d: %s", i, vs)
		if vs.number { // subscript
			vl.def = variables.CreatePMMPVarDecl("<array>", variables.ComplexArray, vl.def)
			subscrCount += 1
		} else { // tag suffix
			vl.def = variables.CreatePMMPVarDecl(vs.text, variables.ComplexSuffix, vl.def)
		}
	}
	T().P("decl", vl.def.GetFullName()).Debugf("full declared type: %v", vl.def)
	// now create variable ref and push onto expression stack
	var subscripts []dec.Decimal = make([]dec.Decimal, subscrCount, subscrCount+1)
	it = vl.suffixes.Iterator()
	for it.Next() { // extract subscripts -> array
		_, vs2 := it.Index(), it.Value().(varsuffix)
		if vs2.number { // subscript
			d, _ := dec.NewFromString(vs2.text)
			subscripts = append(subscripts, d)
		}
	}
	vl.ref = variables.CreatePMMPVarRef(vl.def, nil, subscripts)
	T().P("var", vl.ref.GetName()).Debugf("var ref %v", vl.ref)
}

// Variable parsing: Collect a suffix.
func (vl *varParseListener) ExitSuffix(ctx *grammar.SuffixContext) {
	tag := ctx.TAG().GetText()
	T().Debugf("suffix tag: %s", tag)
	vl.suffixes.Add(varsuffix{tag, false})
}

// Variable parsing: Collect a numeric subscript.
func (vl *varParseListener) ExitSubscript(ctx *grammar.SubscriptContext) {
	d := ctx.DECIMAL().GetText()
	T().Debugf("subscript: %s", d)
	vl.suffixes.Add(varsuffix{d, true})
}
*/
// ----------------------------------------------------------------------

/*
// VariableResolver will find a variable's declaration in a scope tree,
// if present. It will then construct a valid variable reference for
// that declaration.
type VariableResolver interface {
	VariableName() string // full name of variable
	VariableReference(*runtime.ScopeTree) *variables.PMMPVarRef
}

// ParseVariableName will parse a string as a variable's name and
// return a VariableResolver.
func ParseVariableName(v string) (VariableResolver, error) {
	el := &varErrorListener{}
	vtree := parseVariableFromString(v, el)
	if el.err != nil {
		return nil, el.err
	}
	return &varResolver{vtree, "", nil}, el.err
}

type varResolver struct {
	ctx      antlr.RuleContext
	fullname string // variable full name
	varref   *variables.PMMPVarRef
}

func (r *varResolver) VariableReference(sc *runtime.ScopeTree) *variables.PMMPVarRef {
	if r == nil || r.ctx == nil || sc == nil {
		return nil
	}
	if r.varref == nil {
		r.varref, r.fullname = getVarRefFromVarSyntax(r.ctx, sc)
	}
	return r.varref
}

func (r *varResolver) VariableName() string {
	if r.varref == nil {
		r.VariableReference(nil) // provisional call will set fullname
	}
	return r.fullname
}

type varErrorListener struct {
	*antlr.DefaultErrorListener // use default as base class
	err                         error
}

func (el *varErrorListener) SyntaxError(r antlr.Recognizer, sym interface{},
	line, column int, msg string, e antlr.RecognitionException) {
	//
	el.err = fmt.Errorf("[%s|%s] %.44s", strconv.Itoa(line), strconv.Itoa(column), msg)
}
*/
