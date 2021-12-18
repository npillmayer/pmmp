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

func makeToken(state scstate, lexeme string) (tokType, terex.Token) {
	// lmtok := &lex.Token{
	// 	Lexeme: []byte(lexeme),
	// 	Type:   tokenIds[tokcat],
	// 	Value:  nil,
	// }
	toktype := tokval4state[state-accepting_states]
	if toktype == Ident {
		if id, ok := tokenTypeFromLexeme[lexeme]; ok {
			toktype = id
		}
	} else if toktype == Literal {
		toktype = tokType(lexeme[0])
	}
	return toktype, terex.Token{
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
func numberToken(lexeme string, la rune) (tokType, terex.Token) {
	if !unicode.IsLetter(la) {
		return Unsigned, terex.Token{
			Name:    lexeme,
			TokType: int(Unsigned),
			Value:   unsignedValue(lexeme),
		}
	}
	return ScalarMulOp, terex.Token{
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
		if err != nil && (err != io.EOF || r == 0) {
			return 0, nil, l.pos, l.pos
		}
		newstate := nextState(l.state, r)
		if !mustBacktrack(newstate) && newstate != 0 {
			l.match(r)
			l.length += uint64(sz)
		}
		l.state = newstate
		if isAccept(newstate) {
			var t tokType
			if newstate == accept_fraction_bt || newstate == accept_unsigned_bt {
				t, token = numberToken(l.lexeme.String(), r)
				tokval = int(t)
			} else {
				t, token = makeToken(newstate, l.lexeme.String())
				tokval = int(t)
			}
			tracer().Debugf("MetaPost lexer accepting %s", t.String())
			start, length = l.start, l.length
			l.start += length
			l.lexeme.Reset()
			return
		}
	}
}

func (l *lexer) SetErrorHandler(h func(error)) {
	l.errHandler = h
}

func (l *lexer) peek() (r rune, sz int, err error) {
	if l.isEof {
		return 0, 0, io.EOF
	}
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
		r = 1 // TODO this should be 'illegal rune'
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
	// r, sz, err := l.input.ReadRune()
	// l.lookahead.la = r
	// l.lookahead.sz = sz
	// if err == io.EOF {
	// 	l.isEof = true
	// }
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
	state_symtok
	state_dash
	state_ddash
	state_dot
	state_ddot
	state_equals
	state_lt
	state_gt
	state_ast

	accepting_states // do not change sequence, used as a maker
	accept_string
	accept_word
	accept_comment
	accept_literal
	accept_dddash
	accept_dddot
	accept_relop
	accept_primop
	accept_macro_def

	accept_word_bt // do not change sequence
	accept_unsigned_bt
	accept_fraction_bt
	accept_symtok_bt
	accept_minus_bt
	accept_primop_bt
	accept_ddash_bt
	accept_ddot_bt
	max_accepting_states // do not change sequence, used as a marker

	state_err // must be last
)

// callers need to subtract `accepting_states`
var tokval4state = []tokType{
	0, String, Ident, 0, Literal, Join, Join, RelationOp, PrimaryOp, MacroDef,
	Ident, Unsigned, Unsigned, SymTok, PlusOrMinus, PrimaryOp, Join, Join,
}

func mustBacktrack(s scstate) bool {
	return s >= accept_word_bt && s < max_accepting_states
}

func isAccept(s scstate) bool {
	return s > accepting_states && s < max_accepting_states
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
	return charState(s, r)
}

func charState(s scstate, r rune) scstate {
	switch s {
	case state_start:
		switch r {
		case ';', '(', ')', '[', ']', '{', '}', ',':
			return accept_literal
		case '≤', '≥', '≠':
			return accept_relop
		case '=':
			return state_equals
		case '<':
			return state_lt
		case '>':
			return state_gt
		case '#', '&', '@', '$':
			return state_symtok
		case '-':
			return state_dash
		case '*':
			return state_ast
		case '.':
			return state_dot
		}
	case state_equals:
		if r == '=' {
			return accept_relop
		}
		return accept_literal
	case state_lt:
		if r == '=' {
			return accept_relop
		} else if r == '>' {
			return accept_relop
		}
		return accept_literal
	case state_gt:
		if r == '=' {
			return accept_relop
		}
		return accept_literal
	case state_dash:
		switch r {
		case '-':
			return state_ddash
		case '>':
			return accept_macro_def
		default:
			return accept_minus_bt
		}
	case state_dot:
		switch r {
		case '.':
			return state_ddot
		}
		return 0 // single dot is space
	case state_ddot:
		if r == '.' {
			return accept_dddot
		}
		return accept_ddot_bt
	case state_ast:
		if r == '*' {
			return accept_primop
		}
		return accept_primop_bt
	case state_symtok:
		switch r {
		case '#', '@', '$':
			return state_symtok
		}
		return accept_symtok_bt
	}
	return 0
}
