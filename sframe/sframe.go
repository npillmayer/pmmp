package sframe

import (
	"fmt"

	"github.com/npillmayer/gorgo/terex"
)

var GlobalEnvironment *terex.Environment = terex.NewEnvironment("#MP-global", nil)

var GlobalFrameStack ScopeFrameTree

// ---------------------------------------------------------------------------

// For MF/MP, scopes and memory frames collapse to one. There is no static scope tree, as
// scopes are dynamically created for groups.

type DynamicScopeFrame struct {
	ID     int
	env    *terex.Environment
	Parent *DynamicScopeFrame
}

func MakeScopeFrame(id int) DynamicScopeFrame {
	name := fmt.Sprintf("⟨scope %d⟩", id)
	return DynamicScopeFrame{
		ID:  id,
		env: terex.NewEnvironment(name, nil),
	}
}

func (dsf DynamicScopeFrame) Env() *terex.Environment {
	return dsf.env
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
func (scst *ScopeFrameTree) PushNewFrame(id int) *DynamicScopeFrame {
	scp := scst.ScopeTOS
	newsc := MakeScopeFrame(id)
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
