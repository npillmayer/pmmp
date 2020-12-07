package grammar

import (
	"io/ioutil"
	"testing"

	"github.com/npillmayer/gorgo/lr/sppf"
	"github.com/npillmayer/schuko/gtrace"
	"github.com/npillmayer/schuko/tracing"
	"github.com/npillmayer/schuko/tracing/gotestingadapter"
)

func TestParse(t *testing.T) {
	gtrace.SyntaxTracer = gotestingadapter.New()
	teardown := gotestingadapter.RedirectTracing(t)
	defer teardown()
	gtrace.SyntaxTracer.SetTraceLevel(tracing.LevelDebug)
	//
	input := "a.r1"
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
	parsetree := parser.ParseForest()
	tmpfile, err := ioutil.TempFile(".", "parsetree-*.dot")
	if err != nil {
		t.Error("cannot open tmp file for graphviz output")
	}
	sppf.ToGraphViz(parsetree, tmpfile)
	T().Infof("Exported parse tree to %s", tmpfile.Name())
}
