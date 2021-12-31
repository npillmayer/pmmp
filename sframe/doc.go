package sframe

import "github.com/npillmayer/schuko/tracing"

// tracer traces with key 'pmmp.runtime'
func tracer() tracing.Trace {
	return tracing.Select("pmmp.runtime")
}
