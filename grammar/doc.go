package grammar

import (
	"github.com/npillmayer/schuko/tracing"
)

// tracer traces with key 'pmmp.grammar'.
func tracer() tracing.Trace {
	return tracing.Select("pmmp.grammar")
}
