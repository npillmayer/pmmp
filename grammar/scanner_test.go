package grammar

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/npillmayer/gorgo"
	"github.com/npillmayer/schuko/tracing/gotestingadapter"
)

func TestConvertUnsigned(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "pmmp.grammar")
	defer teardown()
	//
	for i, pair := range []struct {
		s string
		v float64
	}{
		{s: "1", v: 1.0},
		{s: "1.0", v: 1.0},
		{s: "1.567", v: 1.567},
		{s: "-1.567", v: -1.567},
		{s: "1/2", v: 0.5},
		{s: "-1/20", v: -0.05},
		{s: "-.5", v: -0.5},
	} {
		if f := unsignedValue(pair.s); f != pair.v {
			t.Errorf("test %d: unsigned %q not recognized: %f", i, pair.s, pair.v)
		}
	}
}

func TestMakeToken(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "pmmp.grammar")
	defer teardown()
	//
	initTokens()
	for i, x := range []struct {
		state scstate
		s     string
		tok   gorgo.TokType
	}{
		{state: accept_symtok, s: "a", tok: SymTok},
		{state: accept_symtok, s: "blabla", tok: SymTok},
		{state: accept_string, s: `"this is a string"`, tok: String},
		{state: accept_symtok, s: "true", tok: NullaryOp},
		{state: accept_symtok, s: "---", tok: Join},
		{state: accept_unsigned_bt, s: "123", tok: Unsigned},
	} {
		if toktype, token := makeToken(x.state, x.s); toktype != x.tok {
			t.Errorf("test %d failed: %v != %v", i, x, token)
		}
	}
}

func TestLexerPeek(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "pmmp.grammar")
	defer teardown()
	//
	input := "test!"
	stream := runeStream{
		reader: bufio.NewReader(strings.NewReader(input)),
		writer: bytes.Buffer{},
	}
	for i := 0; i < 5; i++ {
		r, err := stream.lookahead()
		if err != nil {
			t.Error(err)
		}
		if r != []rune(input)[i] {
			t.Errorf("expected rune #%d to be %#U, is %#U", i, input[i], r)
		}
		stream.match(r)
	}
	_, err := stream.lookahead()
	if err != io.EOF {
		t.Logf("err = %q", err.Error())
		t.Error("expected rune to be 0 and error to be EOF; isn't")
	}
}

func TestLexerState(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "pmmp.grammar")
	defer teardown()
	//
	for i, test := range []struct {
		state scstate
		r     rune
		next  scstate
	}{
		{state: state_start, r: 'a', next: accept_symtok},
		{state: state_comment, r: '\n', next: state_start},
		{state: state_string, r: '\n', next: state_err},
		{state: state_start, r: '\n', next: state_start},
	} {
		csq := catseq{
			c: cat(test.r),
			l: 1,
		}
		if n := next(test.state, csq); n != test.next {
			t.Errorf("test %d failed: %d x %U -> %d expected, was %d", i+1, test.state, test.r, test.next, n)
		}
	}
}

func TestLexerPredicate(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "pmmp.grammar")
	defer teardown()
	//
	if mustBacktrack(state_symtok) {
		t.Errorf("state %d unexpectedly flagged to need backtracking", state_symtok)
	}
	if mustBacktrack(accept_literal) {
		t.Errorf("state %d unexpectedly flagged to need backtracking", accept_literal)
	}
	if !mustBacktrack(accept_unsigned_bt) {
		t.Errorf("state %d unexpectedly flagged to not need backtracking", accept_unsigned_bt)
	}
}

func TestLexerCatSeq(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "pmmp.grammar")
	defer teardown()
	//
	for i, test := range []struct {
		input string
		cat   catcode
		l     int
	}{
		{input: "abc ;", cat: cat0, l: 3},
		{input: "123 ;", cat: cat14, l: 3},
		{input: ">= ;", cat: cat1, l: 2},
		{input: "+-+ ;", cat: cat3, l: 3},
		{input: "();", cat: cat12, l: 1},
	} {
		strm := runeStream{reader: bufio.NewReader(strings.NewReader(test.input))}
		csq, err := nextCategorySequence(&strm)
		if err != nil {
			t.Error(err)
		}
		if csq.l != test.l || csq.c != test.cat {
			t.Errorf("test %d failed: exepected %d|%d, have %d|%d", i+1, test.cat, test.l, csq.c, csq.l)
		}
	}
}

func TestLexerNextToken(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "pmmp.grammar")
	defer teardown()
	//
	initTokens()
	var expect = []gorgo.TokType{
		tokenTypeFromLexeme["begingroup"],
		SymTok, Type, Unsigned, String, Tag, Tag, Unsigned, PrimaryOp, Join, ';', OfOp,
	}
	input := `begingroup @# boolean 1 "hello" a.l 1/23 ** ...;point % ignored`
	lex := NewLexer(bufio.NewReader(strings.NewReader(input)))
	var tokens []gorgo.Token
	for !lex.stream.isEof {
		token := lex.NextToken()
		if token == nil {
			break
		}
		t.Logf("token = %v", token)
		if token.TokType() == EOF {
			break
		}
		tokens = append(tokens, token)
	}
	for i, tok := range expect {
		if i >= len(tokens) {
			t.Fatalf("expected %d tokens, have %d", len(expect), len(tokens))
		}
		if tok != tokens[i].TokType() {
			t.Logf("lexeme #%d = %q", i, tokens[i].Lexeme())
			t.Errorf("expected %d, have %v", tok, tokens[i])
		}
	}
}

func TestLexerMacroDef(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "pmmp.grammar")
	defer teardown()
	//
	input := "-> Hello World enddef !"
	lex := NewLexer(bufio.NewReader(strings.NewReader(input)))
	cat, token, err := lex.storeReplacementText()
	if err != nil {
		t.Fatal(err)
	}
	if cat != MacroDef {
		t.Logf("token = %v", token)
		t.Error("unexpected token category")
		if token.Value() != "-> Hello World " {
			t.Error("unexpected token lexeme")
		}
	}
}

/*
func TestScanner(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "tyse.fonts")
	defer teardown()
	//
	lex, err := Lexer()
	if err != nil {
		t.Errorf(err.Error())
	}
	input := `begingroup @# boolean 1 "hello" a.l 1/23 ** ... ; else % ignored`
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
			t.Logf("token = %-40s with tokval = %d", token, tokval)
			cnt++
		}
	}
	if cnt != 11 {
		t.Errorf("Expected input to be split into 10 tokens, got %d", cnt)
	}
}
*/
