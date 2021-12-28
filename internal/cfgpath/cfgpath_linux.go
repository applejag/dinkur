package cfgpath

import (
	"os"
	"path/filepath"
)

func Path() string {
	const filename = "config.yml"
	if xdgConfig, ok := os.LookupEnv("XDG_CONFIG_HOME"); ok {
		return filepath.Join(xdgConfig, "dinkur", filename)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "dinkur.yml"
	}
	return filepath.Join(home, ".config", "dinkur", filename)
}
