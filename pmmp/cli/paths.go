package cli

// AppPaths is an interface to determine application specific paths for configuration
// and logging/tracing.
type AppPaths interface {
	ConfigDir() string
	LogDir() string
}

// DefaultAppPaths returns an AppPaths instance with platform-dependent defaults
// set, given appTag. appTag is a string specific to a client's application to identify it.
func DefaultAppPaths(appTag string) (AppPaths, error) {
	return appHome(appTag)
}

type appPaths struct {
	tag  string
	home string
}

var _ AppPaths = appPaths{}
