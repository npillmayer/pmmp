package grammar

/*
BSD License

Copyright (c) 2019–21, Norbert Pillmayer

All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions
are met:

1. Redistributions of source code must retain the above copyright
notice, this list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright
notice, this list of conditions and the following disclaimer in the
documentation and/or other materials provided with the distribution.

3. Neither the name of this software nor the names of its contributors
may be used to endorse or promote products derived from this software
without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.  */

import (
	"fmt"
	"sync"

	"github.com/npillmayer/gorgo/lr/scanner"
	"github.com/timtadh/lexmachine"
	lex "github.com/timtadh/lexmachine"
	"github.com/timtadh/lexmachine/machines"
)

// Token values for operators on different grammar levels
const (
	UnaryOp         int = -15
	NullaryOp       int = -16
	PrimaryOp       int = -17
	SecondaryOp     int = -18
	RelationOp      int = -19
	AssignOp        int = -20
	OfOp            int = -21
	UnaryTransform  int = -22
	BinaryTransform int = -23
	Type            int = -24
	Keyword         int = -30
)

// The tokens representing literal one-char lexemes
var literals = []string{
	";", "(", ")", "[", "]", ",",
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
var secOps = []string{`\+`, `\-`, `\+\+`, `\+\-\+`, "or", "intersectionpoint"}
var relOps = []string{
	`=`, `<`, `>`, `≤`, `≥`, `≠`, `<=`, `>=`, `!=`, `<>`,
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

// The keyword tokens
var keywords = []string{ // TODO
	"of",
	`\-\-`, `\.\.`, `\.\.\.`, "tension", "controls",
	"begingroup", "endgroup",
	"def", "vardef",
	"picture", "end",
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
		tokenIds["NUMBER"] = scanner.Float
		tokenIds["STRING"] = scanner.String
		tokenIds["NullaryOp"] = NullaryOp
		tokenIds["UnaryOp"] = UnaryOp
		tokenIds["PrimaryOp"] = PrimaryOp
		tokenIds["SecondaryOp"] = SecondaryOp
		tokenIds["RelationOp"] = RelationOp
		tokenIds["AssignOp"] = AssignOp
		tokenIds["OfOp"] = OfOp
		tokenIds["UnaryTransform"] = UnaryTransform
		tokenIds["BinaryTransform"] = BinaryTransform
		tokenIds["Type"] = Type
		tokenIds["Keyword"] = Keyword
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
		for _, t := range types {
			tokenIds[t] = Type
		}
		for i, k := range keywords {
			tokenIds[k] = Keyword + i
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
		lexer.Add([]byte(`[\+\-]?[0-9]+(\.[0-9]+)?`), makeToken("NUMBER")) // float
		lexer.Add([]byte(`[\+\-]?[0-9]+(\/[0-9]+)?`), makeToken("NUMBER")) // fraction
		lexer.Add([]byte(`([a-z]|[A-Z]|')+`), makeSymbol())
		lexer.Add([]byte(`([a-z]|[A-Z]|')+(\.([a-z|[A-Z]|')+)+`), makeSymbol())
		// lexer.Add([]byte(`([a-z]|[A-Z])+`), makeToken("TAG"))
		// lexer.Add([]byte(`([a-z]|[A-Z])+(\.([a-z|[A-Z])+)+`), makeToken("TAG"))
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
	alltoks = append(alltoks, keywords...)
	T().Debugf("all keywords: %v", alltoks)
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
