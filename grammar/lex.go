package grammar

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

	"github.com/npillmayer/gorgo/terex"
)

// --- Category codes --------------------------------------------------------

type catcode uint8

const (
	cat0  catcode = iota // letter
	cat1                 // <=>:|≤≠≥
	cat2                 // `'´
	cat3                 // +-
	cat4                 // /*\
	cat5                 // !?
	cat6                 // #&@$
	cat7                 // ^~
	cat8                 // [
	cat9                 // ]
	cat10                // {}
	cat11                // .
	cat12                // , ; ( )
	cat13                // "
	cat14                // digit
	cat15                // %
	catNL
	catErr
)

var catcodeTable = []string{
	"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ", // use unicode.IsLetter
	`<=>:|≤≠≥`, "`'", `+-`, `/*\`, `!?`, `#&@$`, `^~`, `[`, `]`, `{}`, `.`, `,;()`, `"`,
	"0123456789", `%`,
}

func cat(r rune) catcode {
	if unicode.IsLetter(r) {
		return cat0
	}
	if unicode.IsDigit(r) {
		return cat14
	}
	for c, cat := range catcodeTable {
		if strings.ContainsRune(cat, r) {
			return catcode(c)
		}
	}
	return catErr
}

// --- Lexer -----------------------------------------------------------------

type lexer struct {
	state      scstate
	stream     runeStream
	errHandler func(error)
}

func NewLexer(reader io.RuneReader) *lexer {
	l := &lexer{}
	l.stream.reader = reader
	return l
}

func (l *lexer) SetErrorHandler(h func(error)) {
	l.errHandler = h
}

func (l *lexer) handleError(err error) {
	if l.errHandler != nil {
		l.errHandler(err)
		return
	}
	tracer().Errorf("MP scanner error: %s", err.Error())
}

// --- Token -----------------------------------------------------------------

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

func (l *lexer) storeReplacementText() (tokType, terex.Token, error) {
	// precondition: we just have read a '->'
	// todo: store text until "enddef" token found
	l.stream.ResetOutput() // drop the '->'
	var r rune
	var err error
	var lexeme string
	//
	edeflen := len("enddef")
	for { // read input until either "enddef" or EOF
		r, err = l.stream.lookahead()
		if err != nil || l.stream.isEof {
			lexeme = l.stream.OutputString()
			break
		}
		l.stream.match(r)
		if r == 'f' { // enddef ends with a 'f' -> now compare backwards
			tracer().Debugf("@ found 'f'")
			lxm := l.stream.writer.Bytes()
			tracer().Debugf("@ lexeme = %q", string(lxm))
			length := len(lxm)
			if length <= edeflen {
				continue
			}
			if isEnddef(lxm[length-edeflen:]) {
				lexeme = string(lxm[:length-edeflen])
				break
			}
		}
	}
	return MacroDef, terex.Token{
		Name:    lexeme,
		TokType: int(ScalarMulOp),
		Value:   lexeme,
	}, err
}

func isEnddef(b []byte) bool {
	var enddef = []byte{'e', 'n', 'd', 'd', 'e', 'f'}
	for i, bb := range b {
		if bb != enddef[i] {
			return false
		}
	}
	return true
}

func (l *lexer) NextToken(expected []int) (tokval int, token interface{}, start, length uint64) {
	if l.stream.isEof {
		return 0, EOF, l.stream.start, l.stream.end
	}
	var r rune
	var err error
	var csq catseq
	for {
		r, err = l.stream.lookahead()
		if err != nil && (err != io.EOF || r == 0) {
			l.handleError(err)
			return 0, nil, l.stream.end, 0
		}
		csq, err = nextCategorySequence(&l.stream, l.state, r)
		if err != nil && (err != io.EOF || r == 0) {
			return 0, nil, l.stream.end, 0
		}
		newstate := next(l.state, csq)
		if !mustBacktrack(newstate) && newstate != 0 {
			l.stream.match(r)
		}
		l.state = newstate
		if isAccept(newstate) {
			var t tokType
			if newstate == accept_fraction_bt || newstate == accept_unsigned_bt {
				t, token = numberToken(l.stream.OutputString(), r)
			} else if newstate == accept_macro_def {
				if t, token, err = l.storeReplacementText(); err != nil {
					tracer().Errorf("MetaPost syntax error: %s", err)
					// TODO make token an error token
					l.handleError(err)
				}
			} else {
				t, token = makeToken(newstate, l.stream.OutputString())
			}
			tokval = int(t)
			tracer().Debugf("MetaPost lexer accepting %s", t.String())
			l.stream.ResetOutput()
			return tokval, token, l.stream.start, l.stream.Span()
		}
	}
}

type token int

const (
	EOF token = iota
)

type scstate int

const (
	state_start scstate = iota
	state_string
	state_symtok
	state_comment
	state_num
	state_frac
	state_denom
	state_macrodef
	// state_dash
	// state_ddash
	// state_dot
	// state_ddot
	// state_equals
	// state_lt
	// state_gt
	// state_ast

	accepting_states // do not change sequence, used as a maker
	accept_string
	// accept_unsigned
	accept_symtok
	accept_comment
	// accept_literal
	// accept_word
	// accept_dddash
	// accept_dddot
	// accept_relop
	// accept_primop
	accept_macro_def

	accept_unsigned_bt // do not change sequence
	accept_fraction_bt
	// accept_word_bt
	// accept_symtok_bt
	// accept_minus_bt
	// accept_primop_bt
	// accept_ddash_bt
	// accept_ddot_bt
	max_accepting_states // do not change sequence, used as a marker

	state_err // must be last
)

// callers need to subtract `accepting_states`
var tokval4state = []tokType{
	0, String, Ident, 0, Literal, Join, Join, RelationOp, PrimaryOp, MacroDef,
	Ident, Unsigned, Unsigned, SymTok, PlusOrMinus, PrimaryOp, Join, Join,
}

func mustBacktrack(s scstate) bool {
	return s >= accept_unsigned_bt && s < max_accepting_states
}

func isAccept(s scstate) bool {
	return s > accepting_states && s < max_accepting_states
}

type catseq struct {
	c catcode // catcode of all runes in this sequence
	l int     // length of sequence in terms of runes
}

func next(s scstate, csq catseq) scstate {
	if s == state_err {
		return s
	}
	switch s {
	case state_start:
		if csq.c <= cat12 {
			return accept_symtok
		}
		switch csq.c {
		case cat13: // "
			return state_string
		case cat14: // digit
			return state_num
		case cat15: // '%'
			return state_comment
		}
	case state_string:
		if csq.c == cat13 {
			return accept_string
		} else if csq.c == catNL {
			return state_err
		}
		return state_string
	case state_num:
		if csq.c == cat11 { // '.'
			return state_frac
		} else if csq.c == cat4 { // '/'
			return state_denom
		}
		return accept_unsigned_bt
	case state_comment:
		if csq.c == catNL {
			return state_start // ignore comments
		}
		return state_comment
	}
	panic("unknown scanner state")
}

// nextCategorySequence reads runes from stream as long as each rune has the same
// category code.
func nextCategorySequence(stream *runeStream, s scstate, r rune) (csq catseq, err error) {
	csq.c = cat(r)
	cc := csq.c
	if csq.c == cat12 { // cat 12 are loners
		csq.l = 1
		return
	}
	for cc == csq.c {
		r, err = stream.lookahead()
		csq.l++
		if err != nil && (err != io.EOF || r == 0) {
			return
		}
		cc = cat(r)
		stream.match(r)
	}
	return
}

// ---------------------------------------------------------------------------

/*
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
*/

/*
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


*/
