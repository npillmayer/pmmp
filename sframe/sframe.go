package sframe

import "github.com/npillmayer/gorgo/terex"

var GlobalEnvironment *terex.Environment = terex.NewEnvironment("#MP-global", nil)

var GlobalFrameStack ScopeFrameTree

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

type TagDeclBase struct {
	terex.Symbol
	Kind TagType
}

func NewTagDecl(name string, kind TagType) TagDecl {
	decl := &TagDeclBase{Kind: kind}
	decl.Symbol.Set("decl", decl) // TODO replace this once generics are there
	return decl
}

func (st TagDeclBase) Name() string {
	return st.Symbol.Name
}

func (st TagDeclBase) TagType() TagType {
	return st.Kind
}

func (st TagDeclBase) IsTag() bool {
	return st.Kind < Spark
}

func (st TagDeclBase) IsSpark() bool {
	return st.Kind >= Spark && st.Kind < TagArray
}

func (st TagDeclBase) IsArray() bool {
	return (st.Kind & TagArray) > 0
}

func (st TagType) AsArray() TagType {
	return st | TagArray
}

// ---------------------------------------------------------------------------

type TagDecl interface {
	Name() string
	TagType() TagType
}

var _ TagDecl = TagDeclBase{}

// ---------------------------------------------------------------------------

// For MF/MP, scopes and memory frames collapse to one. There is no static scope tree, as
// scopes are dynamically created for groups.

type DynamicScopeFrame struct {
	ID     TagDecl
	env    *terex.Environment
	Parent *DynamicScopeFrame
}

func MakeScopeFrame(id TagDecl) DynamicScopeFrame {
	name := "⟨scope⟩"
	if id != nil {
		name = id.Name()
	}
	return DynamicScopeFrame{
		ID:  id,
		env: terex.NewEnvironment(name, nil),
	}
}

func (dsf DynamicScopeFrame) Env() *terex.Environment {
	return dsf.env
}

func (dsf *DynamicScopeFrame) SubType(name string, kind TagType) *DynamicScopeFrame {
	decl := NewTagDecl(name, kind)
	subScope := Derive(decl)
	subScope.Parent = dsf
	return subScope
}

// ScopeFrameTree can be treated as a stack during static analysis, thus we'll be
// building a tree from scopes which are pushed and popped to/from the stack.
//
type ScopeFrameTree struct {
	ScopeBase *DynamicScopeFrame
	ScopeTOS  *DynamicScopeFrame
}

// Current gets the current scope of a stack (TOS).
func (scst *ScopeFrameTree) Current() *DynamicScopeFrame {
	if scst.ScopeTOS == nil {
		panic("attempt to access scope from empty stack")
	}
	return scst.ScopeTOS
}

// Globals gets the outermost scope, containing global symbols.
func (scst *ScopeFrameTree) Globals() *DynamicScopeFrame {
	if scst.ScopeBase == nil {
		panic("attempt to access global scope from empty stack")
	}
	return scst.ScopeBase
}

// PushNewFrame pushes a scope onto the stack of scopes. A scope is constructed, including a symbol table
// for variable declarations.
func (scst *ScopeFrameTree) PushNewFrame(nm TagDecl) *DynamicScopeFrame {
	scp := scst.ScopeTOS
	newsc := MakeScopeFrame(nm)
	newsc.Parent = scp
	if scp == nil { // the new scope is the global scope
		scst.ScopeBase = &newsc // make new scope anchor
		newsc.env = GlobalEnvironment
	} else {
		newsc.env = scp.env
	}
	scst.ScopeTOS = &newsc // new scope now TOS
	tracer().P("scope", newsc.ID).Debugf("pushing new scope")
	return &newsc
}

// PopFrame pops the top-most (recent) scope.
func (scst *ScopeFrameTree) PopFrame() *DynamicScopeFrame {
	if scst.ScopeTOS == nil {
		panic("attempt to pop scope from empty stack")
	}
	sc := scst.ScopeTOS
	tracer().Debugf("popping scope [%s]", sc.ID)
	scst.ScopeTOS = scst.ScopeTOS.Parent
	return sc
}
