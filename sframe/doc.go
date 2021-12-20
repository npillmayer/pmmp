package sframe

import "github.com/npillmayer/schuko/tracing"

// tracer traces with key 'pmmp.scope'
func tracer() tracing.Trace {
	return tracing.Select("pmmp.scope")
}
