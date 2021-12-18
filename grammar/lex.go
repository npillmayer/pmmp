package grammar

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

	"github.com/npillmayer/gorgo/terex"
)

type lexer struct {
	input         io.RuneReader
	state         scstate
	pos           uint64
	start, length uint64 // as bytes index
	lexeme        bytes.Buffer
	lookahead     struct {
		la rune
		sz int
	}
	isEof      bool
	errHandler func(error)
}

func NewLexer(reader io.RuneReader) *lexer {
	l := &lexer{input: reader}
	return l
}

func makeToken(state scstate, lexeme string) terex.Token {
	// lmtok := &lex.Token{
	// 	Lexeme: []byte(lexeme),
	// 	Type:   tokenIds[tokcat],
	// 	Value:  nil,
	// }
	toktype := tokval4state[state-accept_string]
	if toktype == Ident {
		if id, ok := tokenTypeFromLexeme[lexeme]; ok {
			toktype = id
		}
	}
	return terex.Token{
		Name:    lexeme,
		TokType: int(toktype),
		Token:   lexeme, // TODO -> need more info?
	}
}

// TODO
//
// ⟨scalar multiplication op⟩ → +
//     | −
//     | ⟨‘ ⟨number or fraction⟩ ’ not followed by ‘ ⟨add op⟩  ⟨number⟩ ’⟩
//
func (l *lexer) numberToken(lexeme string) terex.Token {
	r, _, err := l.peek()
	if err != nil || !unicode.IsLetter(r) {
		return terex.Token{
			Name:    lexeme,
			TokType: int(Unsigned),
			Value:   unsignedValue(lexeme),
		}
	}
	return terex.Token{
		Name:    lexeme,
		TokType: int(ScalarMulOp),
		Value:   unsignedValue(lexeme),
	}
}

func unsignedValue(s string) float64 {
	var f float64 = 1.0
	if strings.HasPrefix(s, "+") {
		s = s[1:]
	} else if strings.HasPrefix(s, "-") {
		f *= -1.0
		s = s[1:]
	}
	if strings.Contains(s, "/") {
		a := strings.Split(s, "/")
		if len(a) != 2 {
			panic(fmt.Sprintf("malformed fraction: %q", s))
		}
		nom, err1 := strconv.Atoi(a[0])
		denom, err2 := strconv.Atoi(a[1])
		if err1 != nil || err2 != nil {
			panic(fmt.Sprintf("malformed fraction: %q", s))
		}
		f = f * (float64(nom) / float64(denom))
	} else if strings.Contains(s, ".") {
		a, err := strconv.ParseFloat(s, 64)
		if err != nil {
			panic(fmt.Sprintf("malformed fraction: %q", s))
		}
		f = f * a
	}
	return f
}

func (l *lexer) NextToken(expected []int) (tokval int, token interface{}, start, length uint64) {
	var r rune
	var sz int
	var err error
	if l.isEof {
		return 0, EOF, l.pos, l.pos
	}
	for {
		r, sz, err = l.peek()
		if err != nil && err != io.EOF {
			return 0, nil, l.pos, l.pos
		}
		newstate := nextState(l.state, r)
		if !mustBacktrack(newstate) {
			l.match(r)
			l.length += uint64(sz)
		}
		l.state = newstate
		if isAccept(newstate) {
			tokval = int(tokval4state[newstate])
			if newstate == accept_fraction_bt || newstate == accept_unsigned_bt {
				token = l.numberToken(l.lexeme.String())
			} else {
				token = makeToken(newstate, l.lexeme.String())
			}
			start, length = l.start, l.length
			l.start += length
			return
		}
	}
}

func (l *lexer) SetErrorHandler(h func(error)) {
	l.errHandler = h
}

func (l *lexer) peek() (r rune, sz int, err error) {
	if l.lookahead.la != 0 {
		r = l.lookahead.la
		tracer().Debugf("read LA %#U", r)
		sz = l.lookahead.sz
		l.lookahead.la = 0
		return
	}
	r, sz, err = l.input.ReadRune()
	l.lookahead.la = r
	l.lookahead.sz = sz
	if err == io.EOF {
		tracer().Debugf("EOF for MetaPost input")
		l.isEof = true
		return
	} else if err != nil {
		return 0, 0, err
	}
	tracer().Debugf("read rune %#U", r)
	return
}

func (l *lexer) match(r rune) {
	if l.isEof {
		return
	}
	l.lexeme.WriteRune(r)
	if l.lookahead.la != 0 {
		l.lookahead.la = 0
		return
	}
	r, sz, err := l.input.ReadRune()
	l.lookahead.la = r
	l.lookahead.sz = sz
	if err == io.EOF {
		l.isEof = true
	}
}

type token int

const (
	EOF token = iota
)

type scstate int

const (
	state_start scstate = iota
	state_s
	state_w
	state_c
	state_num
	state_ch
	state_frac
	state_denom
	state_macrodef

	accept_string // do not change sequence, used as a maker
	accept_word
	accept_comment

	accept_word_bt // do not change sequence
	accept_unsigned_bt
	accept_fraction_bt

	state_err // must be last
)

var tokval4state = []tokType{
	String, Ident, 0, Ident, Unsigned, Unsigned,
}

func mustBacktrack(s scstate) bool {
	return s >= accept_word_bt && s <= accept_fraction_bt
}

func isAccept(s scstate) bool {
	return s >= accept_string && s <= accept_fraction_bt
}

func nextState(s scstate, r rune) scstate {
	switch s {
	case state_start:
		if unicode.IsLetter(r) {
			return state_w
		}
		if unicode.IsDigit(r) {
			return state_num
		}
		if unicode.IsSpace(r) {
			return state_start
		}
		switch r {
		case '"':
			return state_s
		case '%':
			return state_c
		}
		return charState(s, r)
	case state_w:
		if unicode.IsLetter(r) || r == '\'' {
			return state_w
		}
		if r == '.' {
			return accept_word
		}
		return accept_word_bt
	case state_s:
		if r == '"' {
			return accept_string
		}
		if r == '\n' {
			return state_err
		}
		return state_s
	case state_c:
		if r == '\n' { // comment complete => ignore it
			return state_start
		}
		return state_c
	case state_num:
		if unicode.IsDigit(r) {
			return state_num
		}
		if r == '.' {
			return state_frac
		}
		if r == '/' {
			return state_denom
		}
		return accept_unsigned_bt
	case state_frac:
		if unicode.IsDigit(r) {
			return state_frac
		}
		if r == '.' {
			return state_err
		}
		return accept_unsigned_bt
	case state_denom:
		if unicode.IsDigit(r) {
			return state_denom
		}
		return accept_fraction_bt
	}
	return state_start
}

func charState(s scstate, r rune) scstate {
	switch r {
	case '-':
	}
	return 0
}
