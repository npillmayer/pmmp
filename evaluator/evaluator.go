package evaluator

import (
	"fmt"

	"github.com/npillmayer/arithm/polyn"
	"github.com/npillmayer/gorgo/runtime"
	"github.com/npillmayer/pmmp"
	"github.com/npillmayer/pmmp/variables"
)

// Evaluator is a runtime environment for a PMMP interpreter.
type Evaluator struct {
	*runtime.Runtime                      // interpreter runtime environment
	leq              *polyn.LinEqSolver   // solver for linear equations system
	resolver         map[int]*runtime.Tag // used to resolve variable names from IDs
}

// NewEvaluator creates an evaluating runtime environment.
// It is fully initialized and empty.
func NewEvaluator() *Evaluator {
	ev := &Evaluator{
		Runtime:  runtime.NewRuntimeEnvironment(nil),
		leq:      polyn.CreateLinEqSolver(),
		resolver: make(map[int]*runtime.Tag),
	}
	ev.leq.SetVariableResolver(ev)
	return ev
}

// GetVariableName returns
// the name of a variable, given its ID. Will return the string
// "?nnnn" for capsules.
//
// Interface VariableResolver.
//
func (ev *Evaluator) GetVariableName(id int) string {
	v, ok := ev.resolver[id]
	if !ok {
		return fmt.Sprintf("?%04d", id)
	}
	return v.Name()
}

// IsCapsule is a predicate:
// Is a variable (index) a capsule, i.e., has it gone out of scope?
// The terminus stems from MetaFont (with "whatever" being a prominent
// example for a capsule).
//
// Interface VariableResolver.
//
func (ev *Evaluator) IsCapsule(id int) bool {
	_, found := ev.resolver[id]
	return !found
}

// SetVariableSolved is a notification receiver.
// Set the value of a variable. If the LEQ solves a variable and it becomes
// known, the LEQ will send us this message.
//
// Interface VariableResolver.
//
func (ev *Evaluator) SetVariableSolved(id int, val float64) {
	v, ok := ev.resolver[id]
	if ok { // yes, we know about this variable
		variables.VarFromTag(v).Set(pmmp.FromFloat(val))
	}
}

// EncapsuleVariable drops
// the name of a variable from the variable resolver. The variable itself
// is not dropped, but rather lives on as an anonymous quantity (i.e., a
// capsule) as long as it is part of an equation.
func (ev *Evaluator) EncapsuleVariable(id int32) {
	delete(ev.resolver, int(id))
}

// EncapsuleVarsInMemory makes all variables in a memory frame "capsules".
//
// When a memory frame is popped from the stack, the local variables living
// in the frame have to be made "capsules". This is necessary, because they
// may still be relevant to the LEQ-solver. The LEQ will finally decide
// when to abondon the "zombie" variable.
func (ev *Evaluator) EncapsuleVarsInMemory(mf *runtime.DynamicMemoryFrame) {
	mf.SymbolTable.Each(func(name string, sym *runtime.Tag) {
		vref := variables.VarFromTag(sym)
		tracer().P("var", vref.FullName()).Debugf("encapsule")
		ev.EncapsuleVariable(vref.ID()) // vref is now capsule
	})
}

// AllocateVariableInMemory allocates
// a variable in a given memory frame. Existing variable references in
// this memory frame will be overwritten !
// Clients should probably first call FindVariableReferenceInMemory(vref).
func AllocateVariableInMemory(vref *variables.VarRef,
	mf *runtime.DynamicMemoryFrame) *variables.VarRef {
	//
	mf.SymbolTable.InsertTag(vref.AsTag())
	tracer().P("var", vref.FullName()).Debugf("allocating variable in %s", mf.Name)
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
func (ev *Evaluator) FindVariableReferenceInMemory(vref *variables.VarRef, doAlloc bool) (
	*variables.VarRef, *runtime.DynamicMemoryFrame) {
	//
	if vref.Declaration() == nil {
		tracer().P("var", vref.FullName()).Errorf("attempt to store variable without decl. in memory")
		return vref, nil
	}
	var sym *variables.VarRef
	var memframe *runtime.DynamicMemoryFrame
	tagname := vref.Declaration().FullName()
	tag, scope := ev.ScopeTree.Current().ResolveTag(tagname)
	if tag != nil { // found tag declaration in scope
		memframe = ev.MemFrameStack.FindMemoryFrameForScope(scope)
		varname := vref.Name()
		tracer().P("var", varname).Debugf("var in ? %s", memframe)
		s := memframe.SymbolTable.ResolveTag(varname)
		if s == nil { // no variable ref incarnation => create one
			tracer().P("var", varname).Debugf("not found in memory")
			if doAlloc {
				sym = AllocateVariableInMemory(vref, memframe)
			}
		} else { // already present, return this one
			tracer().P("var", varname).Debugf("variable already present in memory")
			sym = variables.VarFromTag(s)
		}
	} else {
		// this should never happen: we could neither find nor construct a var decl
		panic(fmt.Sprintf("declaration for %s mysteriously vanished...", tagname))
	}
	return sym, memframe
}

// Declare a tag to be of type typ.
//
// If the tag is not declared, insert a new symbol in global scope. If a
// declaration already exists, erase all variables and re-enter a declaration
// (MetaFont semantics). If the tag has been "saved" in the current or in an outer
// scope, make this tag a new undefined symbol.
//
func (ev *Evaluator) Declare(decl *variables.VarDecl) {
	tagname := decl.FullName()
	tag, scope := ev.ScopeTree.Current().ResolveTag(tagname)
	if tag != nil { // already found in scope stack
		tracer().P("tag", tag).Debugf("declare: found tag in scope %s", scope.Name)
		tracer().P("decl", tag).Debugf("variable already declared - re-declaring")
		// Erase all existing variables and re-define symbol
		scope.Tags().InsertTag(decl.AsTag())
	} else { // enter new symbol in global scope
		scope = ev.ScopeTree.Globals()
		scope.Tags().InsertTag(decl.AsTag())
	}
	tracer().P("decl", decl.Name()).Debugf("declared symbol in %s", scope.Name)
}

// Variable creates a variable reference in a memory frame.
// Parameters are the declaration for the variable,
// a value and a flag, indicating if this variable should go to global memory.
// The subscripts parameter is a slice of array-subscripts, if the variable
// declaration is of array (complex) type.
func (ev *Evaluator) Variable(decl *variables.Suffix, value pmmp.Value,
	subscripts []float64, global bool) *variables.VarRef {
	//
	var v *variables.VarRef
	v = variables.CreateVarRef(decl, value, subscripts)
	if global {
		ev.MemFrameStack.Globals().SymbolTable.InsertTag(v.AsTag())
	} else {
		ev.MemFrameStack.Current().SymbolTable.InsertTag(v.AsTag())
	}
	return v
}
