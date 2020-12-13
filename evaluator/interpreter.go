package evaluator

import (
	"fmt"

	"github.com/npillmayer/gorgo/terex"
)

// Interpreter interprets PMMP programs.
type Interpreter struct {
	PC        *terex.GCons // program counter
	IR        instruction  // instruction register
	evaluator *Evaluator   // expression evaluator
	ast       *terex.GCons // program code to execute
	env       *terex.Environment
}

// NewInterpreter creates a new interpreter for the PMMP language.
func NewInterpreter() *Interpreter {
	intp := &Interpreter{
		evaluator: NewEvaluator(),
	}
	return intp
}

// Start expects a list (AST #eof)
func (intp *Interpreter) Start(program *terex.GCons, env *terex.Environment) (err error) {
	intp.ast = program
	intp.env = env
	env.Def("$Evaluator", terex.Elem(intp.evaluator))
	eof := intp.ast.Last()
	if eof == intp.ast {
		T().Errorf("empty program?")
		return fmt.Errorf("no program to execute")
	}
	intp.PC = intp.ast
	//var pc *terex.GCons
	for intp.PC != nil && intp.PC != eof {
		intp.IR, err = intp.fetch(intp.PC)
		if err != nil {
			return
		}
		pc := intp.IR(terex.Elem(intp.PC), env)
		if pc.IsNil() {
			intp.PC = intp.PC.Cdr
		} else {
			intp.PC = pc.AsList()
		}
	}
	return
}

type instruction func(terex.Element, *terex.Environment) terex.Element

func nop(terex.Element, *terex.Environment) terex.Element {
	return terex.Elem(nil)
}

// Fetch the instruction belonging to operator op.
// We do not call the operator directly, but rather search for an operator-symbol
// in the current environment and fetch its 'Call' method.
//
// This indirection allows the AST to be built out of empty operator tags,
// which do nothing. Functionality can be changed by providing different
// operators in the environment (or completely swapping environments with
// different operator sets pre-loaded).
//
// Will return a NOP if operator is not found in environment.
//
func (intp *Interpreter) fetch(astNode *terex.GCons) (instruction, error) {
	e := terex.Elem(astNode.Car)
	if !e.IsAtom() || e.Type() != terex.OperatorType {
		T().Errorf("fetch saw null operator")
		return nop, fmt.Errorf("op fetch saw: %v", astNode.Car)
	}
	op, ok := e.AsAtom().Data.(terex.Operator)
	opname := op.String()
	opsym := intp.env.FindSymbol(opname, true)
	if opsym == nil {
		T().Errorf("Cannot find operation %s", opname)
		return nop, fmt.Errorf("operator not found in env: %v", opname)
	}
	operator, ok := opsym.Value.AsAtom().Data.(terex.Operator)
	if !ok {
		T().Errorf("Cannot call operation %s", opname)
		return nop, fmt.Errorf("Cannot call operation %s", opname)
	}
	return operator.Call, nil
}

// GetEvaluator resolves an Evaluator from an environment. This is to be
// used by operators, which will have been passed an envronment which
// includes a symbol for the calling interpreter's evaluator.
//
func GetEvaluator(env *terex.Environment) *Evaluator {
	esym := env.FindSymbol("$Evaluator", true)
	if esym == nil {
		panic("no evaluator present")
	}
	return esym.Value.AsAtom().Data.(*Evaluator)
}
