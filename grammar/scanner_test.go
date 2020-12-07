package grammar

import (
	"testing"

	"github.com/npillmayer/schuko/gtrace"
	"github.com/npillmayer/schuko/tracing"
	"github.com/npillmayer/schuko/tracing/gotestingadapter"
)

func TestScanner(t *testing.T) {
	gtrace.SyntaxTracer = gotestingadapter.New()
	teardown := gotestingadapter.RedirectTracing(t)
	defer teardown()
	gtrace.SyntaxTracer.SetTraceLevel(tracing.LevelDebug)
	lex, err := Lexer()
	if err != nil {
		t.Errorf(err.Error())
	}
	input := `begingroup boolean 1 "hello" a.l 1/23 ** ... ; % ignored`
	scan, err := lex.Scanner(input)
	if err != nil {
		t.Errorf(err.Error())
	}
	scan.SetErrorHandler(func(e error) {
		t.Error(e)
	})
	done := false
	cnt := 0
	for !done {
		tokval, token, _, _ := scan.NextToken(nil)
		if tokval == -1 {
			done = true
		} else {
			t.Logf("token = %45s with tokval = %d", token, tokval)
			cnt++
		}
	}
	if cnt != 7 {
		t.Errorf("Expected input to be split into 7 tokens, got %d", cnt)
	}
	t.Fail()
}
