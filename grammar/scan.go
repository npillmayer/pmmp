package grammar

/*
License

Governed by a 3-Clause BSD license. License file may be found in the root
folder of this module.

Copyright © 2017–2021 Norbert Pillmayer <norbert@pillmayer.com>

*/

import (
	"fmt"
	"sync"

	"github.com/npillmayer/gorgo"
	"github.com/npillmayer/gorgo/lr/scanner"
)

// Token values for operators on different grammar levels
const (
	EOF             gorgo.TokType = gorgo.TokType(scanner.EOF)
	Tag             gorgo.TokType = -2
	Literal         gorgo.TokType = -3
	String          gorgo.TokType = -4
	SymTok          gorgo.TokType = -9
	Unsigned        gorgo.TokType = -10
	Signed          gorgo.TokType = -11
	UnaryOp         gorgo.TokType = -15
	NullaryOp       gorgo.TokType = -16
	PrimaryOp       gorgo.TokType = -17
	SecondaryOp     gorgo.TokType = -18
	RelationOp      gorgo.TokType = -19
	AssignOp        gorgo.TokType = -20
	OfOp            gorgo.TokType = -21
	UnaryTransform  gorgo.TokType = -22
	BinaryTransform gorgo.TokType = -23
	PlusOrMinus     gorgo.TokType = -24
	Type            gorgo.TokType = -25
	PseudoOp        gorgo.TokType = -26
	Function        gorgo.TokType = -27
	Join            gorgo.TokType = -28
	DrawCmd         gorgo.TokType = -29
	DrawOption      gorgo.TokType = -30
	ScalarMulOp     gorgo.TokType = -31
	MacroDef        gorgo.TokType = -32
	Keyword         gorgo.TokType = -33 // must be the last one => specific keywords will be `Keyword - n`
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
var primOps = []string{`*`, `/`, `**`, "and", "dotprod", "div", "mod"}
var secOps = []string{`++`, `+-+`, "or", "intersectionpoint"}
var sign = []string{`+`, `-`}
var relOps = []string{
	`==`, `<`, `>`, `≤`, `≥`, `≠`, `<=`, `>=`, `<>`,
	`&`, "cutbefore", "cutafter",
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
	`--`, `..`, `...`, `---`,
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
	`[]`,
	"begingroup", "endgroup",
	"picture", "end",
	"tension", "and", "controls", "curl",
	"pickup", "save", "show",
	"def", "vardef", "enddef",
	"expr", "suffix",
	"primary", "secondary", "tertiary",
	"primarydef", "secondarydef", "tertiarydef",
	"if", "fi", "else:", "elseif",
	"for", "endfor", "forsuffixes", "forever", "upto", "downto", "step", "until",
}

// All of the tokens (including literals and keywords)
// var tokens = []string{
// 	"COMMENT", "TAG", "NUMBER", "STRING",
// }

// tokenTypeFromLexeme will be set in initTokens()
var tokenTypeFromLexeme map[string]gorgo.TokType // A map from the token names to their int ids

var initOnce sync.Once // monitors one-time initialization

func initTokens() {
	initOnce.Do(func() {
		tokenTypeFromLexeme = make(map[string]gorgo.TokType)
		// tokenIds["COMMENT"] = scanner.Comment
		// tokenIds["TAG"] = scanner.Ident
		// tokenIds["STRING"] = scanner.String
		// tokenIds["NUMBER"] = scanner.Float
		// tokenIds["Signed"] = Signed
		// tokenIds["Unsigned"] = Unsigned
		// tokenIds["SymTok"] = SymTok
		// tokenIds["NullaryOp"] = NullaryOp
		// tokenIds["UnaryOp"] = UnaryOp
		// tokenIds["PrimaryOp"] = PrimaryOp
		// tokenIds["SecondaryOp"] = SecondaryOp
		// tokenIds["RelationOp"] = RelationOp
		// tokenIds["AssignOp"] = AssignOp
		// tokenIds["OfOp"] = OfOp
		// tokenIds["UnaryTransform"] = UnaryTransform
		// tokenIds["BinaryTransform"] = BinaryTransform
		// tokenIds["PlusOrMinus"] = PlusOrMinus
		// tokenIds["Type"] = Type
		// tokenIds["Function"] = Function
		// tokenIds["Join"] = Join
		// tokenIds["DrawCmd"] = DrawCmd
		// tokenIds["DrawOption"] = DrawOption
		// tokenIds["Keyword"] = Keyword
		// tokenIds["PseudoOp"] = PseudoOp
		for _, lit := range literals {
			r := lit[0]
			tokenTypeFromLexeme[lit] = gorgo.TokType(r)
		}
		for _, op := range nullOps {
			tokenTypeFromLexeme[op] = NullaryOp
		}
		for _, op := range unaryOps {
			tokenTypeFromLexeme[op] = UnaryOp
		}
		for _, op := range primOps {
			tokenTypeFromLexeme[op] = PrimaryOp
		}
		for _, op := range secOps {
			tokenTypeFromLexeme[op] = SecondaryOp
		}
		for _, op := range relOps {
			tokenTypeFromLexeme[op] = RelationOp
		}
		for _, op := range assignOps {
			tokenTypeFromLexeme[op] = AssignOp
		}
		for _, op := range ofOps {
			tokenTypeFromLexeme[op] = OfOp
		}
		for _, tr := range unTransf {
			tokenTypeFromLexeme[tr] = UnaryTransform
		}
		for _, tr := range binTransf {
			tokenTypeFromLexeme[tr] = BinaryTransform
		}
		for _, s := range sign {
			tokenTypeFromLexeme[s] = PlusOrMinus
		}
		for _, t := range types {
			tokenTypeFromLexeme[t] = Type
		}
		for _, f := range funcs {
			tokenTypeFromLexeme[f] = Function
		}
		for _, d := range drawcmd {
			tokenTypeFromLexeme[d] = DrawCmd
		}
		for _, d := range drawopt {
			tokenTypeFromLexeme[d] = DrawOption
		}
		for _, j := range join {
			tokenTypeFromLexeme[j] = Join
			//tokenIds[unescape(j)] = Join
		}
		for i, k := range keywords {
			tokenTypeFromLexeme[k] = Keyword - gorgo.TokType(i)
			//tokenIds[unescape(k)] = Keyword - i
		}
	})
}

// Token returns a token name and its token type.
func Token(t string) (string, gorgo.TokType) {
	id, ok := tokenTypeFromLexeme[t]
	if !ok {
		panic(fmt.Errorf("unknown token: %s", t))
	}
	return t, id
}

// Lexer creates a new lexmachine lexer for the MetaPost language.
/*
func Lexer() (*scanner.LMAdapter, error) {
	initTokens()
	init := func(lexer *lex.Lexer) {
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
*/

/*
func makeToken(s string) lex.Action {
	id, ok := tokenIds[s]
	if !ok {
		panic(fmt.Errorf("unknown token: %s", s))
	}
	return scanner.MakeToken(s, id)
}
*/

/*
func makeSymbol() lex.Action {
	return func(s *lex.Scanner, m *machines.Match) (interface{}, error) {
		lexeme := string(m.Bytes)
		if t, ok := tokenTypeFromLexeme[lexeme]; ok { // is a keyword
			return s.Token(t, lexeme, m), nil
		}
		return s.Token(tokenTypeFromLexeme["TAG"], lexeme, m), nil
	}
}
*/

// S creates a grammar terminal name and the corresponding token value for given
// lexeme. Checks if lexeme is a reserved keyword.
//
// This is used during grammar building.
//
func S(lexeme string) (string, int) {
	if t, ok := tokenTypeFromLexeme[lexeme]; ok { // is a keyword
		return lexeme, int(t)
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
/*
func numberToken(s *lex.Scanner, m *machines.Match) (interface{}, error) {
	lexeme := string(m.Bytes)
	for tc := s.TC; tc < len(s.Text); tc++ { // do not change s.TC
		if unicode.IsSpace(rune(s.Text[tc])) {
			continue
		}
		if unicode.IsLetter(rune(s.Text[tc])) {
			return s.Token(tokenTypeFromLexeme["ScalarMulOp"], lexeme, m), nil
		}
		break
	}
	return s.Token(tokenTypeFromLexeme["NUMBER"], lexeme, m), nil
}
*/

// makeLMToken creates an ad-hoc terminal token.
// Use like this:
//
//     makeLMToken("PrimaryOp", "*")
//
/*
func makeLMToken(tokcat string, lexeme string) *terex.Token {
	lmtok := &lex.Token{
		Lexeme: []byte(lexeme),
		Type:   tokenTypeFromLexeme[tokcat],
		Value:  nil,
	}
	return &terex.Token{
		Name:  tokcat,
		Token: lmtok,
		Value: lexeme,
	}
}
*/
func makeLMToken(tokcat string, lexeme string) gorgo.Token {
	panic("TODO: create tokens during term rewriting")
}

/*
func unescape(s string) string {
	return strings.ReplaceAll(s, `\`, "")
}
*/
