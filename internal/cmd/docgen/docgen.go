package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/dinkur/dinkur/cmd"
	"github.com/spf13/cobra/doc"
)

func main() {
	if len(os.Args) != 2 {
		log.Println("Missing argument: <outputDir>")
		log.Fatalf("Usage: %s <outputDir>", os.Args[0])
	}
	path, err := filepath.Abs(os.Args[1])
	if err != nil {
		log.Fatalln("Error resolving absolute path:", err)
	}
	log.Println("Output dir:", path)
	if err := os.MkdirAll(path, os.ModeDir); err != nil {
		log.Fatalln("Error creating directory:", err)
	}

	cmd.RootCMD.PersistentFlags().Lookup("config").DefValue = "~/.config/dinkur/config.yaml"
	if err := doc.GenMarkdownTree(cmd.RootCMD, path); err != nil {
		log.Fatalln("Error generating markdown tree:", err)
	}
	log.Println("Write complete.")
}
