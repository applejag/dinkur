package cfgpath

import (
	"os"
	"path/filepath"
)

func Path() string {
	appdata, ok := os.LookupEnv("APPDATA")
	if ok {
		return filepath.Join(appdata, "dinkur", "config.yml")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "dinkur.yml"
	}
	return filepath.Join(home, ".dinkur.yml")
}
