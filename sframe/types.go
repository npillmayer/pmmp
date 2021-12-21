package sframe

import (
	"fmt"
	"strings"

	"github.com/npillmayer/gorgo/terex"
)

var idCounter int

func NewID() int {
	idCounter++
	return idCounter
}

type TypeBase struct {
	terex.Symbol
	id         int // for the LEQ every variable has to carry a unique ID
	decl       *TagDeclaration
	subscripts []float64
}

func MakeTypeBase(decl *TagDeclaration, indices ...float64) TypeBase {
	return TypeBase{
		id:         NewID(),
		decl:       decl,
		subscripts: indices,
	}
}

func (tbase TypeBase) ID() int {
	return tbase.id
}

func (tbase TypeBase) Declaration() *TagDeclaration {
	return tbase.decl
}

func (tbase TypeBase) Sym() *terex.Symbol {
	return &tbase.Symbol
}

func (tbase TypeBase) Name() string {
	declname := tbase.decl.Name()
	frags := strings.SplitAfter(declname, "[")
	var inx, fullname string
	if len(frags) > 1 {
		for i, frag := range frags[:len(frags)-1] {
			if i > len(tbase.subscripts) {
				inx = "0"
			} else {
				inx = fmt.Sprintf("%f", tbase.subscripts[i])
			}
			fullname += frag + inx
		}
		fullname += frags[len(frags)-1]
	} else {
		fullname = declname
	}
	return fullname
}

type Variable interface {
	ID() int
	Name() string
	Declaration() *TagDeclaration
	Sym() *terex.Symbol
}

var _ Variable = TypeBase{}

type Numeric struct {
	TypeBase
}

type Pair struct {
	TypeBase
	Parts [2]Numeric
}

type Macro struct {
	TypeBase
	ArgsList    []TagDeclaration
	replacement string
}

func (m Macro) ReplacementText() string {
	return m.replacement
}

type String struct {
	TypeBase
}
