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
	b.LHS("atom").N("number_or_fraction").End()
	b.LHS("number_or_fraction").T("1", 49).End()
	b.LHS("variable").T("x", 120).End()

	g, err := b.Grammar()
	if err != nil {
		T().Errorf("Error creating MetaPost grammar")
		return nil
	}
	g.Dump()
	ga := lr.Analysis(g)
	return ga
}
