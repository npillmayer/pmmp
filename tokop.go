package pmmp

import (
	"fmt"

	"github.com/npillmayer/gorgo"
	"github.com/npillmayer/gorgo/terex"
)

type TokenOperator struct {
	terminalToken gorgo.Token
	call          func(e terex.Element, env *terex.Environment) terex.Element
}

func NewTokenOperator(t gorgo.Token) TokenOperator {
	return TokenOperator{
		terminalToken: t,
		call:          nop,
	}
}

func nop(e terex.Element, env *terex.Environment) terex.Element {
	tracer().Errorf("TokenOperator not to be called")
	return terex.Elem(nil)
}

func (top TokenOperator) String() string {
	// will result in "##<opname>:<op-category>"
	//return "#" + top.Opname() + ":" + top.terminalToken.Name
	return fmt.Sprintf("#%s:%d", top.Opname(), top.Token().TokType())
}

func (top TokenOperator) Opname() string {
	return top.terminalToken.Value().(string)
}

func (top TokenOperator) Token() gorgo.Token {
	return top.terminalToken
}

// Call delegates the operator call to a symbol in the environment.
// The symbol is searched for with the literal value of the operator.
func (top TokenOperator) Call(e terex.Element, env *terex.Environment) terex.Element {
	return callFromEnvironment(top.Opname(), e, env)
}

var _ terex.Operator = &TokenOperator{}

func callFromEnvironment(opname string, e terex.Element, env *terex.Environment) terex.Element {
	opsym := env.FindSymbol(opname, true)
	if opsym == nil {
		tracer().Errorf("Cannot find parsing operation %s", opname)
		return e
	}
	operator, ok := opsym.Value.AsAtom().Data.(terex.Operator)
	if !ok {
		tracer().Errorf("Cannot call parsing operation %s", opname)
		return e
	}
	return operator.Call(e, env)
}
