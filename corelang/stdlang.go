package corelang

import (
	"errors"

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
	env.Defn("def", func(e terex.Element, env *terex.Environment) terex.Element {
		lexeme, toktype, _ := setupFrom(e, env)
		T().Debugf("call of %s/%s", lexeme, toktype)
		args := e.AsList()
		if args.Length() != 2 {
			return ErrorPacker("Wrong number of arguments for def", env)
		}
		return terex.Elem(nil)
	})
}

func ErrorPacker(emsg string, env *terex.Environment) terex.Element {
	T().Errorf(emsg)
	env.Error(errors.New(emsg))
	return terex.Elem(terex.ErrorAtom(emsg))
}

func setupFrom(e terex.Element, env *terex.Environment) (string, string, *evaluator.Evaluator) {
	evaluator := evaluator.GetEvaluator(env)
	if e.First().Type() != terex.OperatorType {
		panic("first element is not an operator")
	}
	op := e.First().AsAtom().Data.(pmmp.TokenOperator)
	return op.Opname(), op.Token().Name, evaluator
}
