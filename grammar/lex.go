package grammar

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

	"github.com/npillmayer/gorgo"
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
	catSpace
	catErr
)

var catcodeTable = []string{
	"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ", // use unicode.IsLetter
	`<=>:|≤≠≥`, "`'", `+-`, `/*\`, `!?`, `#&@$`, `^~`, `[`, `]`, `{}`, `.`, `,;()`, `"`,
	"0123456789", // use unicod.IsDigit
	`%`, "\n\r", " ",
}

func cat(r rune) catcode {
	if unicode.IsLetter(r) {
		return cat0
	}
	if unicode.IsDigit(r) {
		return cat14
	}
	if unicode.IsSpace(r) {
		return catSpace
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
	csq        catseq
	errHandler func(error)
}

func NewLexer(reader io.RuneReader) *lexer {
	l := &lexer{}
	l.stream.reader = &nestedReader{reader: reader}
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

// MPToken is the token type we will return to the parser.
type MPToken struct {
	kind   gorgo.TokType
	lexeme string
	Val    interface{}
	span   gorgo.Span
}

func MakeMPToken(typ gorgo.TokType, lexeme string, value interface{}) MPToken {
	return MPToken{
		kind:   typ,
		lexeme: lexeme,
		Val:    value,
	}
}

func (t MPToken) TokType() gorgo.TokType {
	return t.kind
}

func (t MPToken) Value() interface{} {
	return t.Val
}

func (t MPToken) Lexeme() string {
	return t.lexeme
}

func (t MPToken) Span() gorgo.Span {
	return t.span
}

func makeToken(state scstate, lexeme string) (gorgo.TokType, gorgo.Token) {
	tracer().Debugf("scanner.makeToken state=%d, lexeme=%q", state, lexeme)
	toktype := tokval4state[state-accepting_states]
	if toktype == SymTok {
		if id, ok := tokenTypeFromLexeme[lexeme]; ok {
			toktype = id
		}
	} else if toktype == Literal {
		toktype = gorgo.TokType(lexeme[0])
	}
	return toktype, MPToken{
		lexeme: lexeme,
		kind:   toktype,
	}
}

// TODO
//
// ⟨scalar multiplication op⟩ → +
//     | −
//     | ⟨‘ ⟨number or fraction⟩ ’ not followed by ‘ ⟨add op⟩  ⟨number⟩ ’⟩
//
func numberToken(lexeme string, la rune) (gorgo.TokType, MPToken) {
	if !unicode.IsLetter(la) {
		return Unsigned, MPToken{
			lexeme: lexeme,
			kind:   Unsigned,
			Val:    unsignedValue(lexeme),
		}
	}
	return ScalarMulOp, MPToken{
		lexeme: lexeme,
		kind:   ScalarMulOp,
		Val:    unsignedValue(lexeme),
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

func (l *lexer) storeReplacementText() (gorgo.TokType, MPToken, error) {
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
	return MacroDef, MPToken{
		lexeme: lexeme,
		kind:   ScalarMulOp,
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

func eofToken(pos uint64) gorgo.Token {
	return MPToken{
		kind: EOF,
		span: gorgo.Span{pos, pos},
	}
}

func (l *lexer) NextToken() (token gorgo.Token) {
	if l.stream.isEof {
		return eofToken(l.stream.start)
	}
	var err error
	for {
		if l.csq.l == 0 {
			l.csq, err = nextCategorySequence(&l.stream)
			if err != nil && err != io.EOF {
				// TODO make token an error token
				l.handleError(err)
				return nil
			} else if err == io.EOF {
				return eofToken(l.stream.start)
			}
			tracer().Debugf("scanner category sequence: %v", l.csq)
		}
		newstate := next(l.state, l.csq)
		if !mustBacktrack(newstate) {
			l.csq.l = 0
		}
		l.state = newstate
		tracer().Debugf("scanner new state = %d", newstate)
		if newstate == accept_skip {
			l.stream.ResetOutput()
			l.state = state_start
		} else if isAccept(newstate) {
			token = nil
			if newstate == accept_unsigned || newstate == accept_unsigned_bt {
				r, _ := l.stream.lookahead()
				_, token = numberToken(l.stream.OutputString(), r)
				tracer().Debugf("MetaPost lexer produces :numtoken(%v)", token)
			} else if newstate == accept_macro_def {
				if _, token, err = l.storeReplacementText(); err != nil {
					tracer().Errorf("MetaPost syntax error: %s", err)
					// TODO make token an error token
					l.handleError(err)
				}
				tracer().Debugf("MetaPost lexer stores macro %v", token)
			} else {
				_, token = makeToken(newstate, l.stream.OutputString())
				tracer().Debugf("MetaPost lexer produces :token(%v)", token)
			}
			l.stream.ResetOutput()
			l.state = state_start
			if token == nil {
				panic("scanner token is nil")
			}
			return token
		}
	}
}

// ---------------------------------------------------------------------------

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

	accepting_states // do not change sequence, used as a maker
	accept_skip
	accept_string
	accept_unsigned
	accept_symtok
	accept_literal
	accept_macro_def

	accept_unsigned_bt // do not change sequence
	//accept_fraction_bt
	max_accepting_states // do not change sequence, used as a marker

	state_err // must be last
)

// callers need to subtract `accepting_states` from the input index.
var tokval4state = []gorgo.TokType{
	0, 0, String, Unsigned, SymTok, Literal, MacroDef, Unsigned, Unsigned,
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
		if csq.c == cat11 && csq.l == 1 { // lone single '.' is treated like space
			return accept_skip
		}
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
		return accept_skip
	case state_string:
		if csq.c == cat13 {
			return accept_string
		} else if csq.c == catNL {
			return state_err
		}
		return state_string
	case state_num:
		if csq.c == cat11 && csq.l == 1 { // '.'
			return state_frac
		} else if csq.c == cat4 && csq.l == 1 { // '/'
			return state_denom
		}
		return accept_unsigned_bt
	case state_denom, state_frac:
		if csq.c == cat14 {
			return accept_unsigned
		}
		return state_err
	case state_comment:
		if csq.c == catNL {
			return accept_skip // ignore comments
		}
		return state_comment
	}
	panic(fmt.Sprintf("unknown scanner state: %d", s))
}

// nextCategorySequence reads runes from stream as long as each rune has the same
// category code.
func nextCategorySequence(stream *runeStream) (csq catseq, err error) {
	var r rune
	r, err = stream.lookahead()
	if err != nil && err != io.EOF {
		csq.l = 0
		return csq, fmt.Errorf("scanner cannot read sequence (%w)", err)
	}
	csq.c = cat(r)
	cc := csq.c
	if csq.c == cat12 { // cat 12 are loners
		stream.match(r)
		csq.l = 1
		return
	}
	for cc == csq.c {
		stream.match(r)
		r, err = stream.lookahead()
		csq.l++
		if err != nil && (err != io.EOF || r == 0) {
			return
		}
		cc = cat(r)
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
