package evaluator

import (
	"github.com/npillmayer/schuko/tracing"
)

// tracer traces with key 'pmmp.core'.
func tracer() tracing.Trace {
	return tracing.Select("pmmp.core")
}
