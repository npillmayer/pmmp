//go:build aix || dragonfly || freebsd || (js && wasm) || nacl || linux || netbsd || openbsd || solaris
// +build aix dragonfly freebsd js,wasm nacl linux netbsd openbsd solaris

package cli

import (
	"os"
	"path/filepath"
	"strings"
)

func appHome(appTag string) (a appPaths, err error) {
	a = appPaths{tag: appTag}
	a.home, err = os.UserHomeDir()
	if err != nil {
		a.home = ""
		return
	}
	return
}

func (a appPaths) ConfigDir() string {
	c, err := os.UserConfigDir()
	if err != nil {
		c = filepath.Join(a.home, ".config")
	}
	return filepath.Join(c, strings.ToLower(a.tag))
}

func (a appPaths) LogDir() string {
	c, err := os.UserCacheDir()
	if err != nil {
		c = a.home
	}
	return filepath.Join(c, "logs", strings.ToLower(a.tag))
}
