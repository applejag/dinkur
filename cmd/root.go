// Dinkur the task time tracking utility.
// <https://github.com/dinkur/dinkur>
//
// Copyright (C) 2021 Kalle Fagerberg
// SPDX-FileCopyrightText: 2021 Kalle Fagerberg
// SPDX-License-Identifier: GPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify it under
// the terms of the GNU General Public License as published by the
// Free Software Foundation, either version 3 of the License, or (at your option)
// any later version.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
// FOR A PARTICULAR PURPOSE.  See the GNU General Public License for more
// details.
//
// You should have received a copy of the GNU General Public License along with
// this program.  If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/dinkur/dinkur/internal/cfgpath"
	"github.com/dinkur/dinkur/internal/console"
	"github.com/dinkur/dinkur/pkg/dinkur"
	"github.com/dinkur/dinkur/pkg/dinkurdb"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile = cfgpath.Path()
var flagColor = "auto"
var db dinkur.Client

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:     "dinkur",
	Version: "0.1.0-preview",
	Short:   "The Dinkur CLI",
	Long:    `Through these subcommands you can access your time-tracked tasks.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		switch strings.ToLower(flagColor) {
		case "auto":
			// Do nothing, fatih/color is on auto by default
		case "never":
			color.NoColor = true
		case "always":
			color.NoColor = false
		default:
			console.PrintFatal("Error parsing --color:", fmt.Errorf(`invalid value %q: only "auto", "always", or "never" may be used`, flagColor))
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	db = dinkurdb.NewClient("dinkur.db", dinkurdb.Options{})
	defer db.Close()
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", cfgFile, "config file")
	RootCmd.PersistentFlags().StringVar(&flagColor, "color", flagColor, `colored output: "auto", "always", or "never"`)

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetConfigFile(cfgFile)

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	} else if !errors.As(err, &viper.ConfigFileNotFoundError{}) && !errors.Is(err, os.ErrNotExist) {
		fmt.Fprintln(os.Stderr, "Error reading config:", err)
		os.Exit(1)
	}
}

func connectClientOrExit() {
	if err := db.Connect(); err != nil {
		fmt.Fprintln(os.Stderr, "Error initializing database client:", err)
		os.Exit(1)
	}
}
