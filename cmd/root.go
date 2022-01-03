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
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/dinkur/dinkur/internal/cfgpath"
	"github.com/dinkur/dinkur/internal/console"
	"github.com/dinkur/dinkur/internal/license"
	"github.com/dinkur/dinkur/pkg/dinkur"
	"github.com/dinkur/dinkur/pkg/dinkurclient"
	"github.com/dinkur/dinkur/pkg/dinkurdb"
	"github.com/fatih/color"
	"github.com/iver-wharf/wharf-core/pkg/logger"
	"github.com/iver-wharf/wharf-core/pkg/logger/consolepretty"
	"github.com/mattn/go-colorable"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile       = cfgpath.ConfigPath
	dataFile      = cfgpath.DataPath
	flagDataMkdir = true
	flagColor     = "auto"
	flagClient    = "db"
	flagVerbose   = false

	flagLicenseWarranty   bool
	flagLicenseConditions bool

	c dinkur.Client = &dinkur.NilClient{}

	log = logger.NewScoped("Dinkur")
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:     "dinkur",
	Version: "0.1.0-preview",
	Short:   "The Dinkur CLI",
	Long: license.Header + `
Track how you spend time on your tasks with Dinkur.
`,
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
	Run: func(cmd *cobra.Command, args []string) {
		if flagLicenseWarranty {
			fmt.Println(license.Warranty)
		} else if flagLicenseConditions {
			fmt.Println(license.Conditions)
		} else {
			cmd.Help()
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	defer c.Close()
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig, initLogger)

	RootCmd.SetOut(colorable.NewColorableStdout())
	RootCmd.SetErr(colorable.NewColorableStderr())
	RootCmd.SetUsageTemplate(console.UsageTemplate())

	RootCmd.Flags().BoolVar(&flagLicenseConditions, "license-c", false, "show program's license conditions")
	RootCmd.Flags().BoolVar(&flagLicenseWarranty, "license-w", false, "show program's license warranty")
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", cfgFile, "config file")
	RootCmd.PersistentFlags().StringVar(&dataFile, "data", dataFile, "database file")
	RootCmd.PersistentFlags().BoolVar(&flagDataMkdir, "data-mkdir", flagDataMkdir, "create directory for data if it doesn't exist")
	RootCmd.PersistentFlags().StringVar(&flagColor, "color", flagColor, `colored output: "auto", "always", or "never"`)
	RootCmd.RegisterFlagCompletionFunc("color", colorComplete)
	RootCmd.PersistentFlags().StringVar(&flagClient, "client", flagClient, `Dinkur client: "db" or "grpc"`)
	RootCmd.RegisterFlagCompletionFunc("client", clientComplete)
	RootCmd.PersistentFlags().BoolVarP(&flagVerbose, "verbose", "v", flagVerbose, `enables debug logging`)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetConfigFile(cfgFile)

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Debug().WithString("config", viper.ConfigFileUsed()).Message("Using config file.")
	} else if !errors.As(err, &viper.ConfigFileNotFoundError{}) && !errors.Is(err, os.ErrNotExist) {
		console.PrintFatal("Error reading config:", err)
	}
}

func initLogger() {
	level := logger.LevelInfo
	if flagVerbose {
		level = logger.LevelDebug
	}
	prettyConf := consolepretty.DefaultConfig
	prettyConf.DisableDate = true
	prettyConf.DisableCaller = true
	prettyConf.Writer = colorable.NewColorableStderr()
	logger.AddOutput(level, consolepretty.New(prettyConf))
}

func connectClientOrExit() {
	client, err := connectClient()
	if err != nil {
		console.PrintFatal("Error connecting to client:", err)
	}
	c = client
}

func connectClient() (dinkur.Client, error) {
	switch strings.ToLower(flagClient) {
	case "db":
		log.Debug().Message("Using DB client.")
		dbClient, err := connectToDBClient()
		if err != nil {
			return nil, fmt.Errorf("DB client: %w", err)
		}
		return dbClient, nil
	case "grpc":
		log.Debug().Message("Using gRPC client.")
		grpcClient, err := connectToGRPCClient()
		if err != nil {
			return nil, fmt.Errorf("gRPC client: %w", err)
		}
		return grpcClient, nil
	default:
		return nil, fmt.Errorf(`invalid value %q: only "db" or "grpc" may be used`, flagClient)
	}
}

func connectToGRPCClient() (dinkur.Client, error) {
	c := dinkurclient.NewClient("localhost:59122", dinkurclient.Options{})
	if err := c.Connect(context.Background()); err != nil {
		return nil, err
	}
	if err := c.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("attempting ping: %w", err)
	}
	return c, nil
}

func connectToDBClient() (dinkur.Client, error) {
	c := dinkurdb.NewClient(dataFile, dinkurdb.Options{
		MkdirAll:     flagDataMkdir,
		DebugLogging: flagVerbose,
	})
	return c, c.Connect(context.Background())
}

func colorComplete(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return []string{
		"auto\tuse colored terminal output iff session is interactive (default)",
		"always\talways use colored terminal output; may cause issues when piping output",
		"never\tdisables colored terminal output",
	}, cobra.ShellCompDirectiveDefault
}

func clientComplete(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return []string{
		"grpc\tuse grpc client towards a Dinkur daemon",
		"db\tuse database client directly towards an Sqlite3 file (default)",
	}, cobra.ShellCompDirectiveDefault
}
