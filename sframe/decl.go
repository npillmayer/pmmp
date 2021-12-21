package sframe

import "github.com/npillmayer/gorgo/terex"

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
}

func MakeTagDecl(kind TagType, names ...string) TagDeclaration {
	decl := TagDeclaration{Kind: kind}
	if len(names) == 0 {
		decl.Symbol.Name = "⟨tag-decl⟩"
	} else {
		l := len(names)
		decl.prefixes = make([]string, l)
		var fullname string = names[l-1]
		for i := range names[:l-1] {
			name := names[l-2-i]
			if name != "[]" {
				fullname += "."
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
	return st.Kind
}

func (st TagDeclaration) IsTag() bool {
	return st.Kind < Spark
}

func (st TagDeclaration) IsSpark() bool {
	return st.Kind >= Spark && st.Kind < TagArray
}

func (st TagDeclaration) IsArray() bool {
	return (st.Kind & TagArray) > 0
}

func (st TagType) AsArray() TagType {
	return st | TagArray
}

// ---------------------------------------------------------------------------

/*
type TagDecl interface {
	Name() string
	TagType() TagType
}

var _ TagDecl = TagDeclBase{}
*/

// ---------------------------------------------------------------------------
