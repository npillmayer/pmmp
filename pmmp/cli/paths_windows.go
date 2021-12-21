package cli

import (
	"os"
	"path/filepath"
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
		c = a.home
	}
	return filepath.Join(c, a.tag)
}

func (a appPaths) LogDir() string {
	c, err := os.UserCacheDir()
	if err != nil {
		c = a.home
	}
	return filepath.Join(c, "Logs", a.tag)
}
