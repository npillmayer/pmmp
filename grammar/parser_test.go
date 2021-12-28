package grammar

import (
	"bufio"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/npillmayer/gorgo/lr/sppf"
	"github.com/npillmayer/gorgo/terex"
	"github.com/npillmayer/schuko/tracing"
	"github.com/npillmayer/schuko/tracing/gotestingadapter"
)

func TestParseVariable(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "pmmp.grammar")
	defer teardown()
	//
	parse("a.r1", true, "variable", false, t)
}

func TestParseAtom(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "pmmp.grammar")
	defer teardown()
	//
	parse("true", true, "atom", false, t)
}

func TestParseTertiary1(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "pmmp.grammar")
	defer teardown()
	//
	parse("1+2", true, "tertiary", false, t)
	parse("1+-2", true, "tertiary", false, t)
	parse("1+-1/2a", true, "tertiary", false, t)
	parse("a-(1,2)", true, "tertiary", false, t)
}

func TestParseTertiary2(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "pmmp.grammar")
	defer teardown()
	//
	parse("7 - -a1r", true, "tertiary", false, t) // parser sees: 7 - - a 1 r
}

func TestParseList(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "pmmp.grammar")
	defer teardown()
	//
	parse("a.r1b", true, "variable", false, t)
	parse("true", true, "atom", false, t)
	parse("1/2", true, "atom", false, t)
	parse("xpart z", true, "primary", false, t)
	parse("a * xpart z", true, "secondary", false, t)
	parse("x.r' < -1/4", true, "boolean_expression", false, t)

	parse("numeric p", true, "declaration", false, t)
	parse("pair p[]", true, "declaration", false, t)
	parse("a=1", true, "equation", false, t)
	parse("a=b=5", true, "equation", false, t)
	parse("b = a shifted (1,2)", true, "equation", false, t)
	parse("pair p; p = q;", true, "statement_list", false, t)
	parse("..tension 1.2..", true, "basic_path_join", false, t)
	parse("a = begingroup 5 endgroup", true, "equation", false, t)
	parse("a = begingroup numeric a; 5 endgroup", true, "statement", false, t)
	parse("save a, @$", true, "command", false, t)
	//
	// TODO parse("def a = XXX enddef", true, "macro_definition", false, t)
	// TODO parse("def a(expr x) = XXX enddef;", true, false, t)
}

func TestVariable1AST(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "pmmp.grammar")
	defer teardown()
	//
	compile("a.r1b", "variable", t)
}

func TestAtom1AST(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "pmmp.grammar")
	defer teardown()
	//
	compile("true", "atom", t)
}

func TestAtom2AST(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "pmmp.grammar")
	defer teardown()
	//
	compile("1/2", "atom", t)
}

func TestPrimaryAST(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "pmmp.grammar")
	defer teardown()
	//
	compile("xpart z", "primary", t)
}

func TestSecondaryAST(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "pmmp.grammar")
	defer teardown()
	//
	compile("a * xpart z", "secondary", t)
}

func TestTertiaryAST(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "pmmp.grammar")
	defer teardown()
	//
	compile("a * xpart z - 4", "tertiary", t)
}

func TestExpr1AST(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "pmmp.grammar")
	defer teardown()
	//
	compile("x.r' < -1/4", "boolean_expression", t)
}

func TestExpr2AST(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "pmmp.grammar")
	defer teardown()
	//
	compile("x < (1 + 2)", "boolean_expression", t)
}

func TestVariableAST2(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "pmmp.grammar")
	defer teardown()
	//
	compile("a[7-b]c", "primary", t)
}

func TestPair1(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "pmmp.grammar")
	defer teardown()
	//
	compile("b+(1,3a)", "tertiary", t)
}

func TestInterpolation(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "pmmp.grammar")
	defer teardown()
	//
	compile("-a[1,3b]", "primary", t)
}

func TestOf(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "pmmp.grammar")
	defer teardown()
	//
	compile("point 2 of p", "primary", t)
}

func TestDeclaration1(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "pmmp.grammar")
	defer teardown()
	//
	compile("pair p, q", "declaration", t)
}

func TestDeclaration2(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "pmmp.grammar")
	defer teardown()
	//
	compile("pair p[]r", "declaration", t)
}

func TestEquation1(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "pmmp.grammar")
	defer teardown()
	//
	ast := compile("a=b=c:=5", "statement", t)
	l := terex.Elem(ast).Sublist().AsList().Length() - 1
	t.Logf("eqs = %v", terex.Elem(ast).Sublist())
	if l != 3 {
		t.Errorf("expected sequence of 3 equations, got %d", l)
	}
}

func TestTransformerUn(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "pmmp.grammar")
	defer teardown()
	//
	compile("a shifted (1,2)", "primary", t)
}

func TestTransformerBin(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "pmmp.grammar")
	defer teardown()
	//
	compile("a reflectedabout(b,2)", "primary", t)
}

func TestFuncall(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "pmmp.grammar")
	defer teardown()
	//
	compile("min(1,2,3)", "function_call", t)
}

func TestStatement1(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "pmmp.grammar")
	defer teardown()
	//
	compile("pair p; p = q;", "statement_list", t)
}

func TestPathJoinTension(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "pmmp.grammar")
	defer teardown()
	//
	compile("..tension 1.2..", "basic_join", t)
	compile("..tension 1.2 and 4..", "basic_join", t)
}

func TestPathJoinControls(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "pmmp.grammar")
	defer teardown()
	//
	compile("..controls (1,2)..", "basic_join", t)
	compile("..controls (1,2) and (3,4)..", "basic_join", t)
}

func TestPathAtom(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "pmmp.grammar")
	defer teardown()
	//
	compile("z1", "atom", t) // path
}

func TestPathExpression(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "pmmp.grammar")
	defer teardown()
	//
	compile("z1{curl 1}..controls (1,2)..z2--z3", "path_join", t)
}

func TestGroup(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "pmmp.grammar")
	defer teardown()
	//
	compile("a = begingroup numeric a; 5 endgroup;", "statement_list", t)
}

func TestCommand(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "pmmp.grammar")
	defer teardown()
	//
	compile("save a.r, @$; pickup pencircle; show a;", "statement_list", t)
}

func TestDraw(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "pmmp.grammar")
	defer teardown()
	//
	compile("draw a.r withcolor white withpen pensquare;", "statement_list", t)
}

// ---------------------------------------------------------------------------

func compile(input string, starter string, t *testing.T) *terex.GCons {
	level := tracing.Select("pmmp.grammar").GetTraceLevel()
	tracing.Select("pmmp.grammar").SetTraceLevel(tracing.LevelError)
	tracing.Select("gorgo.lr").SetTraceLevel(tracing.LevelError)
	//
	parser, g := createParser(starter)
	scan := NewLexer(bufio.NewReader(strings.NewReader(input)))
	accept, err := parser.Parse(scan, nil)
	if err != nil || !accept {
		t.Error(err)
		return nil
	}
	tree, retr := parser.ParseForest(), earleyTokenReceiver(parser)
	tracing.Select("pmmp.grammar").SetTraceLevel(tracing.LevelInfo)
	ast, _, err := makeAST(g, tree, retr)
	if err != nil {
		t.Error(err)
	}
	terex.Elem(ast).Dump(tracing.LevelInfo)
	tracing.Select("pmmp.grammar").SetTraceLevel(level)
	if ast == nil {
		t.Errorf("AST is empty")
	}
	return ast
}

func parse(input string, acceptable bool, starter string, dot bool, t *testing.T) {
	level := tracing.Select("pmmp.grammar").GetTraceLevel()
	tracing.Select("pmmp.grammar").SetTraceLevel(tracing.LevelInfo)
	tracing.Select("gorgo.lr").SetTraceLevel(tracing.LevelInfo)
	//
	//scopes := sframe.MakeScopeFrame(77)
	parser, _ := createParser(starter)
	scan := NewLexer(bufio.NewReader(strings.NewReader(input)))
	//
	tracing.Select("pmmp.grammar").SetTraceLevel(tracing.LevelDebug)
	accept, err := parser.Parse(scan, nil)
	t.Logf("accept=%v, input=%s", accept, input)
	if err != nil {
		t.Error(err)
	}
	if acceptable && !accept {
		t.Errorf("expected parser to recognize input, didn't: %q", input)
	} else if !acceptable && accept {
		t.Errorf("expected parser to fail on input, didn't: %q", input)
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
		tmpfile.Close()
	}
	tracing.Select("pmmp.grammar").SetTraceLevel(level)
}
