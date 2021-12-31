package sframe

import (
	"fmt"

	"github.com/npillmayer/gorgo/terex"
)

type TagType uint8

//go:generate stringer -type TagType
const (
	Undefined TagType = iota

	Tag
	TagVardef
	TagNumeric
	TagPair
	TagPath
	TagTransform
	TagString

	Spark
	SparkBuiltin
	SparkMacro
	SparkExpr
	SparkText

	TagArray TagType = 0x01 << 7 // bit flag for tags with array type
)

type TagDeclaration struct {
	terex.Symbol
	Kind     TagType
	prefixes []string
	arraycnt int
}

func MakeTagDecl(kind TagType, names ...string) TagDeclaration {
	decl := TagDeclaration{Kind: kind}
	if len(names) == 0 {
		decl.Symbol.Name = "⟨tag-decl⟩"
	} else {
		l := len(names)
		decl.prefixes = make([]string, l)
		var fullname string
		for i, name := range names {
			if name == "[]" || name == "" {
				if i == 0 {
					panic("tag declaration must have tag as first component (has array)")
				}
				decl.prefixes[i] = "[]"
				decl.arraycnt++
				decl.Kind |= TagArray
			} else {
				decl.prefixes[i] = name
				if i > 0 {
					fullname += "."
				}
			}
			fullname += name
		}
		decl.Symbol.Name = fullname
	}
	decl.Symbol.Set("decl", decl) // TODO replace this once generics are there
	return decl
}

func (st TagDeclaration) Name() string {
	return st.Symbol.Name
}

func (st TagDeclaration) TagType() TagType {
	return st.Kind & 0x03f
}

func (st TagDeclaration) IsTag() bool {
	return st.TagType() < Spark
}

func (st TagDeclaration) IsSpark() bool {
	return st.Kind >= Spark && st.Kind < TagArray
}

func (st TagDeclaration) IsArray() bool {
	return (st.Kind & TagArray) > 0
}

func (st TagDeclaration) Equals(other TagDeclaration) bool {
	return st.Symbol.Name == other.Symbol.Name
}

func (st TagType) AsArray() TagType {
	return st | TagArray
}

// ---------------------------------------------------------------------------

func (st *TagDeclaration) IncarnateVar(indices []float64) Variable {
	var b TypeBase = MakeTypeBase(st, indices...)
	switch st.TagType() {
	case TagNumeric:
		return Numeric{TypeBase: b}
	case TagPair:
		return Pair{TypeBase: b}
	case TagPath:
	case TagTransform:
	case TagString:
		return String{TypeBase: b}
	}
	panic(fmt.Sprintf("not yet implemented: incarnate variable of type %d", st.TagType()))
}
