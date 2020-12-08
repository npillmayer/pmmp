package variables_test

import (
	"testing"

	"github.com/npillmayer/arithm"
	"github.com/npillmayer/gorgo/runtime"
	"github.com/npillmayer/pmmp/variables"
	"github.com/npillmayer/schuko/gtrace"
	"github.com/npillmayer/schuko/tracing"
	"github.com/npillmayer/schuko/tracing/gotestingadapter"
)

// Init the global tracers.

func TestVarDecl1(t *testing.T) {
	symtab := runtime.NewSymbolTable()
	symtab.DefineTag("x")
}

func TestVarDecl2(t *testing.T) {
	gtrace.SyntaxTracer = gotestingadapter.New()
	teardown := gotestingadapter.RedirectTracing(t)
	defer teardown()
	gtrace.SyntaxTracer.SetTraceLevel(tracing.LevelDebug)
	symtab := runtime.NewSymbolTable()
	x := variables.NewVarDecl("x", variables.NumericType)
	symtab.InsertTag(x.Tag())
	variables.CreateSuffix("r", variables.SuffixType, x.AsSuffix())
}

func TestVarDecl3(t *testing.T) {
	gtrace.SyntaxTracer = gotestingadapter.New()
	teardown := gotestingadapter.RedirectTracing(t)
	defer teardown()
	gtrace.SyntaxTracer.SetTraceLevel(tracing.LevelDebug)
	x := variables.NewVarDecl("x", variables.NumericType)
	r := variables.CreateSuffix("r", variables.SuffixType, x.AsSuffix())
	arr := variables.CreateSuffix("<array>", variables.SubscriptType, r)
	variables.CreateSuffix("a", variables.SuffixType, arr)
}

func TestVarDecl4(t *testing.T) {
	gtrace.SyntaxTracer = gotestingadapter.New()
	teardown := gotestingadapter.RedirectTracing(t)
	defer teardown()
	gtrace.SyntaxTracer.SetTraceLevel(tracing.LevelDebug)
	x := variables.NewVarDecl("x", variables.NumericType)
	r := variables.CreateSuffix("r", variables.SuffixType, x.AsSuffix())
	r2 := variables.CreateSuffix("r", variables.SuffixType, x.AsSuffix())
	if r != r2 {
		t.Errorf("Expected suffix not to be re-created, but re-used")
	}
}

func TestVarRef1(t *testing.T) {
	gtrace.SyntaxTracer = gotestingadapter.New()
	teardown := gotestingadapter.RedirectTracing(t)
	defer teardown()
	gtrace.SyntaxTracer.SetTraceLevel(tracing.LevelDebug)
	// build x="hello"
	x := variables.NewVarDecl("x", variables.PathType)
	var v *variables.VarRef = variables.CreateVarRef(x.AsSuffix(), "hello", nil)
	t.Logf("var ref: %v\n", v)
}

func TestVarRef2(t *testing.T) {
	gtrace.SyntaxTracer = gotestingadapter.New()
	teardown := gotestingadapter.RedirectTracing(t)
	defer teardown()
	gtrace.SyntaxTracer.SetTraceLevel(tracing.LevelDebug)
	// build x.r=1
	x := variables.NewVarDecl("x", variables.NumericType)
	r := variables.CreateSuffix("r", variables.SuffixType, x.AsSuffix())
	var v *variables.VarRef = variables.CreateVarRef(r, 1, nil)
	t.Logf("var ref: %v\n", v)
}

func TestVarRef3(t *testing.T) {
	gtrace.SyntaxTracer = gotestingadapter.New()
	teardown := gotestingadapter.RedirectTracing(t)
	defer teardown()
	gtrace.SyntaxTracer.SetTraceLevel(tracing.LevelDebug)
	// build x[7]=1
	x := variables.NewVarDecl("x", variables.NumericType)
	arr := variables.CreateSuffix("<[]>", variables.SubscriptType, x.AsSuffix())
	subs := []float64{7.0}
	var v *variables.VarRef = variables.CreateVarRef(arr, 1.0, subs)
	t.Logf("var ref: %v\n", v)
}

func TestVarRefPair1(t *testing.T) {
	gtrace.SyntaxTracer = gotestingadapter.New()
	teardown := gotestingadapter.RedirectTracing(t)
	defer teardown()
	gtrace.SyntaxTracer.SetTraceLevel(tracing.LevelDebug)
	// build x.r=(1,2)
	x := variables.NewVarDecl("x", variables.PairType)
	r := variables.CreateSuffix("r", variables.SuffixType, x.AsSuffix())
	var v *variables.VarRef = variables.CreateVarRef(r, arithm.P(1, 2), nil)
	if !v.IsPair() {
		t.Errorf("Expected v to be a pair variable, but is not")
	}
	t.Logf("var ref: %v\n", v)
	ypart := v.YPart()
	t.Logf("ypart of v = %v", ypart.String())
	if ypart.Value != 2.0 {
		t.Errorf("Expected ypart of x.r to have value=2, has not")
	}
	t.Fail()
}
