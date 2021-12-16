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
	teardown := gotestingadapter.QuickConfig(t, "tyse.fonts")
	defer teardown()
	//
	parse("a.r1", false, t)
}

func TestParseAtom(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "tyse.fonts")
	defer teardown()
	//
	parse("true", false, t)
}

func TestParseSecondary(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "tyse.fonts")
	defer teardown()
	//
	parse("1+2", false, t)
	parse("1+-2", false, t)
	parse("1+-1/2a", false, t)
	parse("a-(1,2)", false, t)
}

func TestParseTertiary(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "tyse.fonts")
	defer teardown()
	//
	parse("7 - -a1r", false, t) // parser sees: 7 - - a 1 r
}

func TestParseList(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "tyse.fonts")
	defer teardown()
	//
	// parse("a.r1b", false, t)
	// parse("true", false, t)
	// parse("1/2", false, t)
	// parse("xpart z", false, t)
	// parse("a * xpart z", false, t)
	// parse("x.r' < -1/4", false, t)

	// parse("numeric p;", false, t)
	// parse("pair p[];", false, t)
	// parse("a=1;", false, t)
	// parse("a=b=5;", false, t)
	// parse("b = a shifted (1,2)", false, t)
	// parse("pair p; p = q;", false, t)
	// parse("..tension 1.2..;", false, t)
	// parse("a = begingroup 5 endgroup;", false, t)
	// parse("a = begingroup numeric a; 5 endgroup;", false, t)
	// parse("save a, @$;", false, t)

	//parse("def a = XXX enddef;", true, t)
	parse("def a(expr x) = XXX enddef;", true, t)
}

func TestVariableAST1(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "tyse.fonts")
	defer teardown()
	//
	compile("a.r1b", t)
}

func TestAtom1AST(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "tyse.fonts")
	defer teardown()
	//
	compile("true", t)
}

func TestAtom2AST(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "tyse.fonts")
	defer teardown()
	//
	compile("1/2", t)
}

func TestPrimaryAST(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "tyse.fonts")
	defer teardown()
	//
	compile("xpart z", t)
}

func TestSecondaryAST(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "tyse.fonts")
	defer teardown()
	//
	compile("a * xpart z", t)
}

func TestTertiaryAST(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "tyse.fonts")
	defer teardown()
	//
	compile("a * xpart z - 4", t)
}

func TestExpr1AST(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "tyse.fonts")
	defer teardown()
	//
	compile("x.r' < -1/4", t)
}

func TestExpr2AST(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "tyse.fonts")
	defer teardown()
	//
	compile("x < (1 + 2)", t)
}

func TestVariableAST2(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "tyse.fonts")
	defer teardown()
	//
	compile("a[7-b]c", t)
}

func TestPair1(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "tyse.fonts")
	defer teardown()
	//
	compile("b+(1,3a)", t)
}

func TestInterpolation(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "tyse.fonts")
	defer teardown()
	//
	compile("-a[1,3b]", t)
}

func TestOf(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "tyse.fonts")
	defer teardown()
	//
	compile("point 2 of p", t)
}

func TestDeclaration1(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "tyse.fonts")
	defer teardown()
	//
	compile("pair p, q", t)
}

func TestDeclaration2(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "tyse.fonts")
	defer teardown()
	//
	compile("pair p[]r", t)
}

func TestEquation1(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "tyse.fonts")
	defer teardown()
	//
	ast := compile("a=b=c:=5", t)
	l := terex.Elem(ast).Sublist().AsList().Length() - 1
	t.Logf("eqs = %v", terex.Elem(ast).Sublist())
	if l != 3 {
		t.Errorf("expected sequence of 3 equations, got %d", l)
	}
	t.Fail()
}

func TestTransformerUn(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "tyse.fonts")
	defer teardown()
	//
	compile("a shifted (1,2)", t)
}

func TestTransformerBin(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "tyse.fonts")
	defer teardown()
	//
	compile("a reflectedabout(b,2)", t)
}

func TestFuncall(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "tyse.fonts")
	defer teardown()
	//
	compile("min(1,2,3)", t)
}

func TestStatement1(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "tyse.fonts")
	defer teardown()
	//
	compile("pair p; p = q;", t)
}

func TestPathJoinTension(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "tyse.fonts")
	defer teardown()
	//
	//compile("..tension 1.2..;", t)
	compile("..tension 1.2 and 4..;", t)
}

func TestPathJoinControls(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "tyse.fonts")
	defer teardown()
	//
	//compile("..controls (1,2)..;", t)
	compile("..controls (1,2) and (3,4)..;", t)
	t.Fail()
}

func TestPathAtom(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "tyse.fonts")
	defer teardown()
	//
	compile("z1;", t)
}

func TestPathExpression(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "tyse.fonts")
	defer teardown()
	//
	compile("z1{curl 1}..controls (1,2)..z2--z3;", t)
}

func TestGroup(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "tyse.fonts")
	defer teardown()
	//
	compile("a = begingroup numeric a; 5 endgroup;", t)
}

func TestCommand(t *testing.T) {
	teardown := gotestingadapter.RedirectTracing(t)
	defer teardown()
	//
	compile("save a.r, @$; pickup pencircle; show a;", t)
	t.Fail()
}

func TestDraw(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "tyse.fonts")
	defer teardown()
	//
	compile("draw a.r withcolor white withpen pensquare;", t)
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
			tracer().Infof("Exported parse tree to %s", tmpfile.Name())
		}
	}
}
