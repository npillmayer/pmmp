package grammar

/*
License

Governed by a 3-Clause BSD license. License file may be found in the root
folder of this module.

Copyright © 2017–2021 Norbert Pillmayer <norbert@pillmayer.com>

*/

import (
	"fmt"
	"strings"
	"sync"
	"unicode"

	"github.com/npillmayer/gorgo/lr/scanner"
	"github.com/npillmayer/gorgo/terex"
	"github.com/timtadh/lexmachine"
	lex "github.com/timtadh/lexmachine"
	"github.com/timtadh/lexmachine/machines"
)

// Token values for operators on different grammar levels
const (
	SymTok          int = -9
	Unsigned        int = -10
	Signed          int = -11
	UnaryOp         int = -15
	NullaryOp       int = -16
	PrimaryOp       int = -17
	SecondaryOp     int = -18
	RelationOp      int = -19
	AssignOp        int = -20
	OfOp            int = -21
	UnaryTransform  int = -22
	BinaryTransform int = -23
	PlusOrMinus     int = -24
	Type            int = -25
	PseudoOp        int = -26
	Function        int = -27
	Join            int = -28
	DrawCmd         int = -29
	DrawOption      int = -30
	Keyword         int = -32
)

// The tokens representing literal one-char lexemes
var literals = []string{
	";", "(", ")", "[", "]", "{", "}", ",", "=",
}
var types = []string{
	"boolean", "cmycolor", "color", "numeric", "pair", "path", "pen",
	"picture", "rgbcolor", "string", "transform",
}
var unaryOps = []string{ // TODO
	"abs", "angle",
	//
	"xpart", "ypart", "yellowpart",
}
var nullOps = []string{
	"false", "normaldeviate", "nullpen", "nullpicture",
	"pencircle", "true", "whatever",
}
var primOps = []string{`\*`, `\/`, `\*\*`, "and", "dotprod", "div", "mod"}
var secOps = []string{`\+\+`, `\+\-\+`, "or", "intersectionpoint"}
var sign = []string{`\+`, `\-`}
var relOps = []string{
	`==`, `<`, `>`, `≤`, `≥`, `≠`, `<=`, `>=`, `!=`, `<>`,
	`\&`, "cutbefore", "cutafter",
}
var assignOps = []string{`:=`, `←`}
var ofOps = []string{
	"arctime", "direction", "directiontime", "directionpoint", "penoffset",
	"point", "postcontrol", "precontrol", "subpath", "substring",
}
var unTransf = []string{
	"rotated", "scaled", "shifted", "slanted", "transformed",
	"xscaled", "yscaled", "zscaled",
}
var binTransf = []string{
	"reflectedabout", "reflectedaround",
}
var funcs = []string{
	"min", "max", "incr", "decr",
}
var join = []string{
	`--`, `\.\.`, `\.\.\.`, `---`,
}
var drawcmd = []string{
	"draw", "fill", "filldraw", "undraw", "unfill", "unfilldraw",
	"drawarrow", "drawdblarrow", "cutdraw",
}
var drawopt = []string{
	"withcolor", "withrgbcolor", "withcmykcolor",
	"withgreyscale", "withpen", "dashed",
}

// The keyword tokens
var keywords = []string{ // TODO
	"of",
	`\[\]`,
	"begingroup", "endgroup",
	"picture", "end",
	"tension", "and", "controls", "curl",
	"pickup", "save", "show",
	"def", "vardef", "enddef", "XXX",
	"expr", "suffix",
	"primary", "secondary", "tertiary",
	"primarydef", "secondarydef", "tertiarydef",
}

// All of the tokens (including literals and keywords)
var tokens = []string{
	"COMMENT", "TAG", "NUMBER", "STRING",
}

// tokenIds will be set in initTokens()
var tokenIds map[string]int // A map from the token names to their int ids

var initOnce sync.Once // monitors one-time initialization

func initTokens() {
	initOnce.Do(func() {
		tokenIds = make(map[string]int)
		tokenIds["COMMENT"] = scanner.Comment
		tokenIds["TAG"] = scanner.Ident
		tokenIds["STRING"] = scanner.String
		tokenIds["NUMBER"] = scanner.Float
		tokenIds["Signed"] = Signed
		tokenIds["Unsigned"] = Unsigned
		tokenIds["SymTok"] = SymTok
		tokenIds["NullaryOp"] = NullaryOp
		tokenIds["UnaryOp"] = UnaryOp
		tokenIds["PrimaryOp"] = PrimaryOp
		tokenIds["SecondaryOp"] = SecondaryOp
		tokenIds["RelationOp"] = RelationOp
		tokenIds["AssignOp"] = AssignOp
		tokenIds["OfOp"] = OfOp
		tokenIds["UnaryTransform"] = UnaryTransform
		tokenIds["BinaryTransform"] = BinaryTransform
		tokenIds["PlusOrMinus"] = PlusOrMinus
		tokenIds["Type"] = Type
		tokenIds["Function"] = Function
		tokenIds["Join"] = Join
		tokenIds["DrawCmd"] = DrawCmd
		tokenIds["DrawOption"] = DrawOption
		tokenIds["Keyword"] = Keyword
		tokenIds["PseudoOp"] = PseudoOp
		for _, lit := range literals {
			r := lit[0]
			tokenIds[lit] = int(r)
		}
		for _, op := range nullOps {
			tokenIds[op] = NullaryOp
		}
		for _, op := range unaryOps {
			tokenIds[op] = UnaryOp
		}
		for _, op := range primOps {
			tokenIds[op] = PrimaryOp
		}
		for _, op := range secOps {
			tokenIds[op] = SecondaryOp
		}
		for _, op := range relOps {
			tokenIds[op] = RelationOp
		}
		for _, op := range assignOps {
			tokenIds[op] = AssignOp
		}
		for _, op := range ofOps {
			tokenIds[op] = OfOp
		}
		for _, tr := range unTransf {
			tokenIds[tr] = UnaryTransform
		}
		for _, tr := range binTransf {
			tokenIds[tr] = BinaryTransform
		}
		for _, s := range sign {
			tokenIds[s] = PlusOrMinus
		}
		for _, t := range types {
			tokenIds[t] = Type
		}
		for _, f := range funcs {
			tokenIds[f] = Function
		}
		for _, d := range drawcmd {
			tokenIds[d] = DrawCmd
		}
		for _, d := range drawopt {
			tokenIds[d] = DrawOption
		}
		for _, j := range join {
			tokenIds[j] = Join
			tokenIds[unescape(j)] = Join
		}
		for i, k := range keywords {
			tokenIds[k] = Keyword - i
			tokenIds[unescape(k)] = Keyword - i
		}
	})
}

// Token returns a token name and its value.
func Token(t string) (string, int) {
	id, ok := tokenIds[t]
	if !ok {
		panic(fmt.Errorf("unknown token: %s", t))
	}
	return t, id
}

// Lexer creates a new lexmachine lexer for the MetaPost language.
func Lexer() (*scanner.LMAdapter, error) {
	initTokens()
	init := func(lexer *lexmachine.Lexer) {
		lexer.Add([]byte(`%[^\n]*\n?`), scanner.Skip) // skip comments
		lexer.Add([]byte(`\"[^"]*\"`), makeToken("STRING"))
		//lexer.Add([]byte(`[\+\-]\d+(\.\d+)?`), makeToken("Signed")) // float
		//lexer.Add([]byte(`[\+\-]\d+(/\d+)?`), makeToken("Signed"))  // fraction
		lexer.Add([]byte(`\d+(\.\d+)?`), makeToken("Unsigned")) // float
		lexer.Add([]byte(`\d+(/\d+)?`), makeToken("Unsigned"))  // fraction
		lexer.Add([]byte(`([a-zA-Z']|')+`), makeSymbol())
		lexer.Add([]byte(`([a-zA-Z'])+(\.([a-zA-Z'])+)+`), makeSymbol())
		lexer.Add([]byte(`[#&@$]+`), makeToken("SymTok"))
		lexer.Add([]byte(`( |\t|\n|\r)+`), scanner.Skip) // skip whitespace
	}
	alltoks := append(nullOps, unaryOps...)
	alltoks = append(alltoks, primOps...)
	alltoks = append(alltoks, secOps...)
	alltoks = append(alltoks, relOps...)
	alltoks = append(alltoks, assignOps...)
	alltoks = append(alltoks, ofOps...)
	alltoks = append(alltoks, unTransf...)
	alltoks = append(alltoks, binTransf...)
	alltoks = append(alltoks, types...)
	alltoks = append(alltoks, sign...)
	alltoks = append(alltoks, funcs...)
	alltoks = append(alltoks, drawcmd...)
	alltoks = append(alltoks, join...)
	alltoks = append(alltoks, keywords...)
	tracer().Debugf("all keywords: %v", alltoks)
	adapter, err := scanner.NewLMAdapter(init, literals, alltoks, tokenIds)
	if err != nil {
		return nil, err
	}
	return adapter, nil
}

func makeToken(s string) lexmachine.Action {
	id, ok := tokenIds[s]
	if !ok {
		panic(fmt.Errorf("unknown token: %s", s))
	}
	return scanner.MakeToken(s, id)
}

func makeSymbol() lexmachine.Action {
	return func(s *lex.Scanner, m *machines.Match) (interface{}, error) {
		lexeme := string(m.Bytes)
		if t, ok := tokenIds[lexeme]; ok { // is a keyword
			return s.Token(t, lexeme, m), nil
		}
		return s.Token(tokenIds["TAG"], lexeme, m), nil
	}
}

// S creates a grammar terminal name and the corresponding token value for given
// lexeme. Checks if lexeme is a reserved keyword.
func S(lexeme string) (string, int) {
	if t, ok := tokenIds[lexeme]; ok { // is a keyword
		return lexeme, t
	}
	panic(fmt.Sprintf("did not find token value for lexeme '%s'", lexeme))
	//return lexeme, tokenIds["TAG"] // this should not happen
}

// TODO:
//
// ⟨scalar multiplication op⟩ → +
//     | −
//     | ⟨‘ ⟨number or fraction⟩ ’ not followed by ‘ ⟨add op⟩  ⟨number⟩ ’⟩
//
func numberToken(s *lex.Scanner, m *machines.Match) (interface{}, error) {
	lexeme := string(m.Bytes)
	for tc := s.TC; tc < len(s.Text); tc++ { // do not change s.TC
		if unicode.IsSpace(rune(s.Text[tc])) {
			continue
		}
		if unicode.IsLetter(rune(s.Text[tc])) {
			return s.Token(tokenIds["ScalarMulOp"], lexeme, m), nil
		}
		break
	}
	return s.Token(tokenIds["NUMBER"], lexeme, m), nil
}

// makeLMToken creates an ad-hoc terminal token.
// Use like this:
//
//     makeLMToken("PrimaryOp", "*")
//
func makeLMToken(tokcat string, lexeme string) *terex.Token {
	lmtok := &lexmachine.Token{
		Lexeme: []byte(lexeme),
		Type:   tokenIds[tokcat],
		Value:  nil,
	}
	return &terex.Token{
		Name:  tokcat,
		Token: lmtok,
		Value: lexeme,
	}
}

func unescape(s string) string {
	return strings.ReplaceAll(s, `\`, "")
}
