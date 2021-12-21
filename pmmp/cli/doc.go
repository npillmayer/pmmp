package cli

import "github.com/npillmayer/schuko/tracing"

// tracer traces with key 'pmmp.cli'
func tracer() tracing.Trace {
	return tracing.Select("pmmp.cli")
}
