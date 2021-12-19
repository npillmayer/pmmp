package grammar

import (
	"bufio"
	"io"
	"strings"
	"testing"

	"github.com/npillmayer/gorgo/terex"
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
		tok   tokType
	}{
		{state: accept_word_bt, s: "a", tok: Ident},
		{state: accept_word_bt, s: "blabla", tok: Ident},
		{state: accept_string, s: `"blabla"`, tok: String},
		{state: accept_word_bt, s: "true", tok: NullaryOp},
		{state: accept_word_bt, s: "min", tok: Function},
	} {
		if toktype, _ := makeToken(x.state, x.s); toktype != x.tok {
			t.Errorf("test %d failed: %v", i, x)
		}
	}
}

func TestLexerPeek(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "pmmp.grammar")
	defer teardown()
	//
	input := "test!"
	lex := NewLexer(bufio.NewReader(strings.NewReader(input)))
	for i := 0; i < 5; i++ {
		r, _, err := lex.peek()
		if err != nil {
			t.Error(err)
		}
		if r != []rune(input)[i] {
			t.Errorf("expected rune #%d to be %#U, is %#U", i, input[i], r)
		}
		lex.match(r)
	}
	r, _, err := lex.peek()
	if r != 0 || err != io.EOF {
		t.Logf("r = %#U, err = %q", r, err.Error())
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
		{state: 0, r: 'a', next: state_w},
		{state: state_w, r: 'b', next: state_w},
		{state: state_c, r: '\n', next: 0},
		{state: state_w, r: '\n', next: accept_word_bt},
		{state: 0, r: '\n', next: 0},
	} {
		if n := nextState(test.state, test.r); n != test.next {
			t.Errorf("test %d failed: %d x %U -> %d expected, was %d", i, test.state, test.r, test.next, n)
		}
	}
}

func TestLexerPredicate(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "pmmp.grammar")
	defer teardown()
	//
	if mustBacktrack(state_w) {
		t.Errorf("state %d unexpectedly flagged to need backtracking", state_w)
	}
	if mustBacktrack(accept_comment) {
		t.Errorf("state %d unexpectedly flagged to need backtracking", accept_comment)
	}
	if !mustBacktrack(accept_word_bt) {
		t.Errorf("state %d unexpectedly flagged to not need backtracking", accept_word_bt)
	}
}

func TestLexerNextToken(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "pmmp.grammar")
	defer teardown()
	//
	initTokens()
	var expect = []tokType{
		tokenTypeFromLexeme["begingroup"],
		SymTok, Type, Unsigned, String, Ident, Ident, Unsigned, PrimaryOp, Join, ';', OfOp,
	}
	input := `begingroup @# boolean 1 "hello" a.l 1/23 ** ... ; point % ignored`
	lex := NewLexer(bufio.NewReader(strings.NewReader(input)))
	var cats []tokType
	var lexemes []string
	for !lex.isEof {
		cat, token, _, _ := lex.NextToken(nil)
		t.Logf("cat = %s, token = %v", tokType(cat), token)
		cats = append(cats, tokType(cat))
		if token == nil {
			lexemes = append(lexemes, "<nil>")
		} else {
			lexemes = append(lexemes, token.(terex.Token).Name)
		}
	}
	for i, tok := range expect {
		if i >= len(cats) {
			t.Fatalf("expected %d tokens, have %d", len(expect), len(cats))
		}
		if tok != cats[i] {
			t.Logf("lexeme #%d = %q", i, lexemes[i])
			t.Errorf("expected %q, have %q", tok, cats[i])
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
		t.Logf("cat = %s, token = %v", cat.String(), token)
		t.Error("unexpected token category")
		if token.Value != "-> Hello World " {
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
