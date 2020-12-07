package grammar

import (
	"github.com/npillmayer/gorgo/lr"
	"github.com/npillmayer/schuko/gtrace"
	"github.com/npillmayer/schuko/tracing"
)

// T traces to the global syntax tracer.
func T() tracing.Trace {
	return gtrace.SyntaxTracer
}

func MakeMetaPostGrammar() *lr.LRAnalysis {
	b := lr.NewGrammarBuilder("MetaPost")

	b.LHS("primary").N("atom").End()
	b.LHS("atom").N("variable").End()
	b.LHS("atom").T(S("NUMBER")).End()
	b.LHS("variable").T(S("TAG")).N("suffix").End()
	b.LHS("suffix").Epsilon()
	b.LHS("suffix").N("suffix").N("subscript").End()
	b.LHS("suffix").N("suffix").T(S("TAG")).End()
	b.LHS("subscript").T(S("NUMBER")).End()

	g, err := b.Grammar()
	if err != nil {
		T().Errorf("Error creating MetaPost grammar")
		return nil
	}
	g.Dump()
	ga := lr.Analysis(g)
	return ga
}
