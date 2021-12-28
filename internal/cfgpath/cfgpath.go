//go:build !windows && !linux
// +build !windows,!linux

package cfgpath

import (
	"os"
	"path/filepath"
)

func Path() string {
	configDir, err := os.UserHomeDir()
	if err != nil {
		return "dinkur.yml"
	}
	return filepath.Join(configDir, ".dinkur.yml")
}
