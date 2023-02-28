// Dinkur the task time tracking utility.
// <https://github.com/dinkur/dinkur>
//
// Copyright (C) 2021 Kalle Fagerberg
// SPDX-FileCopyrightText: 2021 Kalle Fagerberg
// SPDX-License-Identifier: GPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify it
// under the terms of the GNU General Public License as published by the
// Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for
// more details.
//
// You should have received a copy of the GNU General Public License along
// with this program.  If not, see <http://www.gnu.org/licenses/>.

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

	cmd.RootCmd.PersistentFlags().Lookup("sqlite.path").DefValue = "~/.local/share/dinkur/dinkur.db"
	if err := doc.GenMarkdownTree(cmd.RootCmd, path); err != nil {
		log.Fatalln("Error generating markdown tree:", err)
	}
	log.Println("Write complete.")
}
