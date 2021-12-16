package evaluator_test

import (
	"testing"

	"github.com/npillmayer/gorgo/terex"
	"github.com/npillmayer/pmmp"
	"github.com/npillmayer/pmmp/corelang"
	"github.com/npillmayer/pmmp/evaluator"
	"github.com/npillmayer/schuko/gtrace"
	"github.com/npillmayer/schuko/tracing/gotestingadapter"
)

func TestSetup(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "tyse.fonts")
	defer teardown()
	//
	intp := evaluator.NewInterpreter()
	if intp == nil {
		t.Errorf("error creating interpreter")
	}
	_, err := intp.Start(nil, nil)
	if err != nil {
		if err != evaluator.ErrNoProgramToExecute {
			t.Errorf("expected empty-input-error, but got %v", err)
		}
	} else {
		t.Error("expected Start to fail due to empty input, but didn't")
	}
}

func TestStart1(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "tyse.fonts")
	defer teardown()
	//
	intp := evaluator.NewInterpreter()
	ast := terex.Atomize(terex.Cons(wrap("+", "SecondaryOp"), nil))
	input := terex.Cons(ast, terex.Cons(wrap("#eof", "EOF"), nil))
	_, err := intp.Start(input, nil)
	if err != nil {
		t.Errorf("error executing trivial program: %v", err)
	}
}

func TestSubtraction(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "tyse.fonts")
	defer teardown()
	//
	intp := evaluator.NewInterpreter()
	operands := terex.Cons(num(3), terex.Cons(num(2), nil))
	ast := terex.Atomize(terex.Cons(wrap("-", "SecondaryOp"), operands))
	input := terex.Cons(ast, terex.Cons(wrap("#eof", "EOF"), nil))
	stdlang := corelang.LoadStandardLanguage()
	gtrace.SyntaxTracer.Debugf(stdlang.Dump())
	r, err := intp.Start(input, stdlang)
	if err != nil {
		t.Errorf("error executing trivial program: %v", err)
	}
	if v, ok := r.Car.Data.(pmmp.Value); !ok {
		t.Errorf("expected return value to be of type pmmp.Value")
	} else if !v.Self().IsNumeric() || v.Self().AsNumeric().Polynomial().GetConstantValue() != 1.0 {
		t.Errorf("expected interpreter to return numeric constant 1.0, is %v", v.Self())
	}
}

// --- Helpers ---------------------------------------------------------------

type testTok struct {
	name  string
	Value string
}

func wrap(opname string, cat string) terex.Atom {
	tok := testTok{
		name:  opname,
		Value: opname,
	}
	t := terex.Token{
		Name:    cat,
		TokType: int(terex.OperatorType),
		Token:   tok,
		Value:   opname,
	}
	return terex.Atomize(pmmp.NewTokenOperator(&t))
}

func num(f float64) terex.Atom {
	return terex.Atomize(f)
}
