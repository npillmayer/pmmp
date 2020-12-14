package corelang

import (
	"errors"
	"fmt"

	"github.com/npillmayer/gorgo/terex"
	"github.com/npillmayer/pmmp"
	"github.com/npillmayer/pmmp/evaluator"
)

func LoadStandardLanguage() *terex.Environment {
	env := terex.NewEnvironment("pmmplang", nil)
	defineExprOps(env)
	return env
}

func defineExprOps(env *terex.Environment) {
	env.Defn("SecondaryOp", func(e terex.Element, env *terex.Environment) terex.Element {
		lexeme, toktype, _, thread := setupFrom(e, env)
		T().Debugf("call of %s/%s", lexeme, toktype)
		errelem, _, argv := args(e, 2, env)
		if !errelem.IsNil() {
			return errelem
		}
		e1 := thread.FetchDecodeExecute(terex.Elem(argv.Nth(1)))
		e2 := thread.FetchDecodeExecute(terex.Elem(argv.Nth(2)))
		if iserr(e1) || iserr(e2) {
			return ErrorPacker("error converting arguments", env)
		}
		v1, v2 := value(e1), value(e2)
		var err error
		var v pmmp.Value
		switch lexeme {
		case "+":
			//r, err := num1.Self().Plus(num2)
			v, err = v1.Self().Minus(v2) // TODO Plus
		case "-":
			v, err = v1.Self().Minus(v2)
		default:
			return ErrorPacker(fmt.Sprintf("unknown secondary operator %s", lexeme), env)
		}
		if err != nil {
			//
		}
		T().Debugf("%v %s %v = %s", v1.Self(), lexeme, v2.Self(), v.Self())
		return terex.Elem(v)
	})
}

func args(e terex.Element, n int, env *terex.Environment) (terex.Element, int, *terex.GCons) {
	argc := e.AsList().Length() - 1
	if n >= 0 && argc != n {
		return ErrorPacker("Wrong number of arguments for operator", env), 0, nil
	}
	l := e.AsList().Cdr
	return terex.Elem(nil), argc, l
}

func ErrorPacker(emsg string, env *terex.Environment) terex.Element {
	T().Errorf(emsg)
	env.Error(errors.New(emsg))
	return terex.Elem(terex.ErrorAtom(emsg))
}

func iserr(e terex.Element) bool {
	return e.Type() == terex.ErrorType
}

func setupFrom(e terex.Element, env *terex.Environment) (
	string, string, *evaluator.Evaluator, *evaluator.Thread) {
	eval := evaluator.GetEvaluator(env)
	th := evaluator.GetThread(env)
	if e.First().Type() != terex.OperatorType {
		panic("first element is not an operator")
	}
	op := e.First().AsAtom().Data.(pmmp.TokenOperator)
	return op.Opname(), op.Token().Name, eval, th
}

func value(e terex.Element) pmmp.Value {
	if e.Type() == terex.NumType {
		return pmmp.FromFloat(e.AsAtom().Data.(float64))
	}
	if e.Type() == terex.UserType {
		d := e.AsAtom().Data
		if v, ok := d.(pmmp.Value); ok {
			return v
		}
	}
	panic(fmt.Sprintf("expected value, got %v", e))
	//return pmmp.Numeric{} // TODO other types + vars
	// check for error: list, nil, …
}
