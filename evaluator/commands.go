package evaluator

import (
	"bytes"
	"fmt"

	"github.com/npillmayer/gorgo/runtime"
	"github.com/npillmayer/pmmp"
	"github.com/npillmayer/pmmp/variables"
)

// WhateverDeclaration is the global declaration of 'whatever', used to
// instantiate anonymous whatever-variables.
var WhateverDeclaration *variables.VarDecl

// Counter for 'whatever' anonymous variables.
var whateverCounter int64

// Whatever creates an anonymous variable. In MetaFont this is a macro, but
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

/*
Assign is a variable assignment.

   assignment : lvalue ASSIGN numtertiary

(1) Retract lvalue from the resolver's table (make a capsule)

(3) Unset the value of lvalue

(3) Re-incarnate lvalue (get a new ID for it)

(4) If type is numeric or pair: Create equation on expression stack,
else assign a path value to a path variable.
*/
func (ev *Evaluator) Assign(lvalue *variables.VarRef, e pmmp.Value) {
	varname := lvalue.Name()
	oldserial := lvalue.ID
	T().P("var", varname).Debugf("assignment of lvalue #%d", oldserial)
	ev.EncapsuleVariable(lvalue.ID())
	vref, mf := ev.FindVariableReferenceInMemory(lvalue, false)
	vref.Set(nil) // now lvalue is unset / unsolved
	T().P("var", varname).Debugf("unset in %v", mf)
	vref.Reincarnate()
	T().P("var", vref.Name()).Debugf("new lvalue incarnation #%d", vref.ID)
	if vref.Type() == pmmp.PathType {
		//vref.Set(e.Other) // TODO Value of type path
	} else { // create linear equation
		// TODO
		//exprStack(rt).EquateTOS2OS()  // construct equation
	}
}

// Save a tag within a group. The tag will be restored at the end of the
// group. Save-commands within global scope will be ignored.
//
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
func (ev *Evaluator) Endgroup() {
	mf := ev.PopScopeAndMemory()
	ev.EncapsuleVarsInMemory(mf)
}

// PopScopeAndMemory decreases the grouping level.
// We pop the topmost scope and topmost memory frame. This happens after
// a group is left.
//
// Returns the previously topmost memory frame.
//
func (ev *Evaluator) PopScopeAndMemory() *runtime.DynamicMemoryFrame {
	hidden := ev.ScopeTree.PopScope()
	hidden.Name = "(hidden)"
	mf := ev.MemFrameStack.PopMemoryFrame()
	if mf.Scope != hidden {
		T().P("mem", mf.Name).Errorf("groups out of sync?")
	}
	return mf
}

// --- Show commands ---------------------------------------------------------

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
			if vref.Declaration().AsTag() == sym {
				s := fmt.Sprintf("%s = %s\n", vref.FullName(), vref.ValueString())
				b.WriteString(s)
			}
		}
	}
	return b.String()
}
