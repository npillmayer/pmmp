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

func TestParseTertiary(t *testing.T) {
	gtrace.SyntaxTracer = gotestingadapter.New()
	teardown := gotestingadapter.RedirectTracing(t)
	defer teardown()
	gtrace.SyntaxTracer.SetTraceLevel(tracing.LevelDebug)
	parse("7 - -a1r", false, t) // parser sees: 7 - - a 1 r
}

func TestParseList(t *testing.T) {
	gtrace.SyntaxTracer = gotestingadapter.New()
	teardown := gotestingadapter.RedirectTracing(t)
	defer teardown()
	gtrace.SyntaxTracer.SetTraceLevel(tracing.LevelDebug)
	parse("a.r1b", false, t)
	parse("true", false, t)
	parse("1/2", false, t)
	parse("xpart z", false, t)
	parse("a * xpart z", false, t)
	parse("x.r' < -1/4", false, t)
}

func TestVariableAST(t *testing.T) {
	gtrace.SyntaxTracer = gotestingadapter.New()
	teardown := gotestingadapter.RedirectTracing(t)
	defer teardown()
	compile("a.r1b", t)
}

func TestAtom1AST(t *testing.T) {
	gtrace.SyntaxTracer = gotestingadapter.New()
	teardown := gotestingadapter.RedirectTracing(t)
	defer teardown()
	compile("true", t)
}

func TestAtom2AST(t *testing.T) {
	gtrace.SyntaxTracer = gotestingadapter.New()
	teardown := gotestingadapter.RedirectTracing(t)
	defer teardown()
	compile("1/2", t)
}

func TestPrimaryAST(t *testing.T) {
	gtrace.SyntaxTracer = gotestingadapter.New()
	teardown := gotestingadapter.RedirectTracing(t)
	defer teardown()
	compile("xpart z", t)
}

func TestSecondaryAST(t *testing.T) {
	gtrace.SyntaxTracer = gotestingadapter.New()
	teardown := gotestingadapter.RedirectTracing(t)
	defer teardown()
	compile("a * xpart z", t)
}

func TestTertiaryAST(t *testing.T) {
	gtrace.SyntaxTracer = gotestingadapter.New()
	teardown := gotestingadapter.RedirectTracing(t)
	defer teardown()
	compile("a * xpart z - 4", t)
}

func TestExpr1AST(t *testing.T) {
	gtrace.SyntaxTracer = gotestingadapter.New()
	teardown := gotestingadapter.RedirectTracing(t)
	defer teardown()
	compile("x.r' < -1/4", t)
}

func TestExpr2AST(t *testing.T) {
	gtrace.SyntaxTracer = gotestingadapter.New()
	teardown := gotestingadapter.RedirectTracing(t)
	defer teardown()
	compile("x < (1 + 2)", t)
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
	if ast == nil {
		t.Errorf("AST is empty")
	}
	return ast
}

func parse(input string, dot bool, t *testing.T) {
	gtrace.SyntaxTracer.SetTraceLevel(tracing.LevelInfo)
	parser := createParser()
	scan, _ := mpLexer.Scanner(input)
	gtrace.SyntaxTracer.SetTraceLevel(tracing.LevelDebug)
	accept, err := parser.Parse(scan, nil)
	t.Logf("accept=%v, input=%s", accept, input)
	if err != nil {
		t.Error(err)
	}
	if !accept {
		t.Errorf("No accept. Not a valid MetaPost expression")
	}
	if err == nil && accept && dot {
		parsetree := parser.ParseForest()
		tmpfile, err := ioutil.TempFile(".", "parsetree-*.dot")
		if err != nil {
			t.Error("cannot open tmp file for graphviz output")
		} else {
			sppf.ToGraphViz(parsetree, tmpfile)
			T().Infof("Exported parse tree to %s", tmpfile.Name())
		}
	}
}
