package vm

import "github.com/npillmayer/schuko/tracing"

// tracer traces with key 'pmmp.vm'
func tracer() tracing.Trace {
	return tracing.Select("pmmp.vm")
}
