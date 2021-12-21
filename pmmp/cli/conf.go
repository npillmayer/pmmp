package cli

import (
	"strings"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/npillmayer/pmmp"
	"github.com/npillmayer/schuko/schukonf/koanfadapter"
	"github.com/npillmayer/schuko/tracing"
	"github.com/npillmayer/schuko/tracing/gologadapter"
	"github.com/npillmayer/schuko/tracing/trace2go"
)

// loadConfig is a callback function used by cobra's initialization mechanism.
// Unfortunately we're not allowed a return value.
func loadConfig() {
	k := koanf.New(".") // '.' is hierarchy delimiter
	// We locate pmmp configuration with an application-key of 'PMMP' and
	// use NestedText-format (nt) for config-files
	konf := koanfadapter.New(k, "PMMP", []string{"nt"})
	konf.InitDefaults()
	if err := mergeFlags(konf); err != nil {
		tracing.Errorf(err.Error())
		pmmp.Exit(1)
	}
	//fmt.Printf("@@@ tracing.dest is now %q\n", konf.GetString("tracing.destination"))
	if err := configureTracing(konf); err != nil {
		tracing.Errorf(err.Error())
		pmmp.Exit(1)
	}
	pmmp.Configuration = k // push the configuratoin to app-global scope
}

func mergeFlags(konf *koanfadapter.KConf) error {
	flags := rootCmd.PersistentFlags()
	err := konf.Koanf().Load(posflag.Provider(flags, ".", konf.Koanf()), nil)
	if err != nil {
		return err
	}
	if logname := konf.GetString("logfile"); logname != "" && logname != "stderr" {
		if strings.Contains(logname, ":/") {
			konf.Set("tracing.destination", logname)
		} else {
			konf.Set("tracing.destination", "file://"+logname)
		}
	}
	return err
}

func configureTracing(konf *koanfadapter.KConf) error {
	if a := konf.GetString("tracing.adapter"); a != "" && a != "go" {
		tracing.Errorf("tracing adapter type '%s' currently not supported", a)
	}
	konf.Set("tracing.adapter", "go") // use Go builtin logging facilities
	tracing.Infof("searching for trace redirection")
	paths := locateLogFile()
	if dest := konf.GetString("tracing.destination"); dest != "" {
		//fmt.Printf("@@@ dest/pre  = %q\n", dest)
		if !strings.Contains(dest, ":") && paths.ConfigDir() != "" {
			dest = "file://" + paths.ConfigDir() + "/" + dest
			//fmt.Printf("@@@ dest/post = %q\n", dest)
			konf.Set("tracing.destination", dest)
		}
		// tracing.Infof("opening tracing destination %q\n", dest)
		// if d, err := tracing.Destination(dest); err != nil {
		// 	return fmt.Errorf("re-directing trace output failed: %w", err)
		// } else {
		// 	fmt.Printf("@@@ d = %T\n", d)
		// 	trace2go.Root().SetOutput(d)
		// 	tracing.Infof(rootCmd.Long)
		// }
	}
	tracing.RegisterTraceAdapter("go", gologadapter.GetAdapter(), false)
	if err := trace2go.ConfigureRoot(konf, "trace", trace2go.ReplaceTracers(true)); err != nil {
		return err
	}
	tracing.SetTraceSelector(trace2go.Selector())
	tracing.Infof(rootCmd.Long)
	return nil
}

func locateLogFile() AppPaths {
	paths, err := DefaultAppPaths("PMMP")
	if err != nil {
		tracing.Errorf("cannot configure paths: %v", err)
	}
	//fmt.Printf("@ config path: %q\n", paths.ConfigDir())
	//fmt.Printf("@ log path   : %q\n", paths.LogDir())
	return paths
}
