package sframe

type TypeBase struct {
	decl  TagDecl
	index float64
}

func (tbase TypeBase) Declaration() TagDecl {
	return tbase.decl
}

func (tbase TypeBase) Index() float64 {
	return tbase.index
}

type Variable interface {
	Declaration() TagDecl
	Index() float64
}

var _ Variable = TypeBase{}

type Numeric struct {
	TypeBase
	value float64
}

type Pair struct {
	TypeBase
	pair [2]float64
}

type Macro struct {
	TypeBase
	ArgsList    []TagDeclBase
	replacement string
}

func (m Macro) ReplacementText() string {
	return m.replacement
}

type String struct {
	TypeBase
	value string
}

func Derive(sym TagDecl) *DynamicScopeFrame {
	scope := &DynamicScopeFrame{}
	scope.ID = sym
	return scope
}
