/*
Package corelang implements core commands for DSLs dealing with arithmetic
expressions, pairs and paths. It borrows from MetaFont/MetaPost,
as described in the accompanying ANTLR grammar file.

Language Features

This package includes functions for numerous language features common to
MetaFont-/MetaPost-derivated DSLs.

The implementation is tightly coupled to the ANTLR V4 parser generator.
ANTLR is a great tool and I see no use in being independent from it.

Lua Scripting

This package also includes the support for Lua scripting. The DSLs stemming
from this language core are Lua-enabled by default.

For further information please refer to types Scripting and LuaVarRef.

___________________________________________________________________________

License

Governed by a 3-Clause BSD license. License file may be found in the root
folder of this module.

Copyright © 2017–2021 Norbert Pillmayer <norbert@pillmayer.com>

*/
package corelang

import (
	"github.com/npillmayer/schuko/tracing"
)

// tracer traces with key 'pmmp.core'.
func tracer() tracing.Trace {
	return tracing.Select("pmmp.core")
}
