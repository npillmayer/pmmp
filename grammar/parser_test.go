package grammar

import (
	"io/ioutil"
	"testing"

	"github.com/npillmayer/gorgo/lr/sppf"
	"github.com/npillmayer/gorgo/terex"
	"github.com/npillmayer/schuko/gtrace"
	"github.com/npillmayer/schuko/tracing"
	"github.com/npillmayer/schuko/tracing/gotestingadapter"
)

func TestParseVariable(t *testing.T) {
	gtrace.SyntaxTracer = gotestingadapter.New()
	teardown := gotestingadapter.RedirectTracing(t)
	defer teardown()
	gtrace.SyntaxTracer.SetTraceLevel(tracing.LevelDebug)
	parse("a.r1", false, t)
}

func TestParseAtom(t *testing.T) {
	gtrace.SyntaxTracer = gotestingadapter.New()
	teardown := gotestingadapter.RedirectTracing(t)
	defer teardown()
	gtrace.SyntaxTracer.SetTraceLevel(tracing.LevelDebug)
	parse("true", false, t)
}

func TestParseSecondary(t *testing.T) {
	gtrace.SyntaxTracer = gotestingadapter.New()
	teardown := gotestingadapter.RedirectTracing(t)
	defer teardown()
	gtrace.SyntaxTracer.SetTraceLevel(tracing.LevelDebug)
	parse("a*b", false, t)
}

func TestVariableAST(t *testing.T) {
	gtrace.SyntaxTracer = gotestingadapter.New()
	teardown := gotestingadapter.RedirectTracing(t)
	defer teardown()
	ast := compile("a.r1b", t)
	if ast == nil {
		t.Fail()
	}
}

func TestAtomAST(t *testing.T) {
	gtrace.SyntaxTracer = gotestingadapter.New()
	teardown := gotestingadapter.RedirectTracing(t)
	defer teardown()
	ast := compile("true", t)
	if ast == nil {
		t.Fail()
	}
}

func TestPrimaryAST(t *testing.T) {
	gtrace.SyntaxTracer = gotestingadapter.New()
	teardown := gotestingadapter.RedirectTracing(t)
	defer teardown()
	ast := compile("xpart z", t)
	if ast == nil {
		t.Fail()
	}
}

func TestSecondaryAST(t *testing.T) {
	gtrace.SyntaxTracer = gotestingadapter.New()
	teardown := gotestingadapter.RedirectTracing(t)
	defer teardown()
	ast := compile("a * xpart z", t)
	//ast := compile("a * b", t)
	if ast == nil {
		t.Fail()
	}
	t.Fail()
}

// ---------------------------------------------------------------------------

func compile(input string, t *testing.T) *terex.GCons {
	gtrace.SyntaxTracer.SetTraceLevel(tracing.LevelError)
	tree, retr, err := Parse(input)
	if err != nil {
		t.Error(err)
	}
	gtrace.SyntaxTracer.SetTraceLevel(tracing.LevelInfo)
	ast, _, err := AST(tree, retr)
	if err != nil {
		t.Error(err)
	}
	terex.Elem(ast).Dump(tracing.LevelInfo)
	return ast
}

func parse(input string, dot bool, t *testing.T) {
	gtrace.SyntaxTracer.SetTraceLevel(tracing.LevelInfo)
	parser := createParser()
	scan, _ := mpLexer.Scanner(input)
	gtrace.SyntaxTracer.SetTraceLevel(tracing.LevelInfo)
	accept, err := parser.Parse(scan, nil)
	t.Logf("accept=%v, input=%s", accept, input)
	if err != nil {
		t.Error(err)
	}
	if !accept {
		t.Errorf("No accept. Not a valid MetaPost expression")
	}
	parsetree := parser.ParseForest()
	if dot {
		tmpfile, err := ioutil.TempFile(".", "parsetree-*.dot")
		if err != nil {
			t.Error("cannot open tmp file for graphviz output")
		}
		sppf.ToGraphViz(parsetree, tmpfile)
		T().Infof("Exported parse tree to %s", tmpfile.Name())
	}
}
