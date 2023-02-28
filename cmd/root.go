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

package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/dinkur/dinkur/internal/console"
	"github.com/dinkur/dinkur/internal/license"
	"github.com/dinkur/dinkur/pkg/config"
	"github.com/dinkur/dinkur/pkg/dinkur"
	"github.com/dinkur/dinkur/pkg/dinkurclient"
	"github.com/dinkur/dinkur/pkg/dinkurdb"
	"github.com/fatih/color"
	"github.com/iver-wharf/wharf-core/v2/pkg/logger"
	"github.com/iver-wharf/wharf-core/v2/pkg/logger/consolejson"
	"github.com/iver-wharf/wharf-core/v2/pkg/logger/consolepretty"
	"github.com/mattn/go-colorable"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfg     = config.Default
	cfgFile string

	rootCtx     = context.Background()
	rootCtxDone func()

	flagVerbose = false

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
Track how you spend time on your entries with Dinkur.
`,
	SilenceErrors: true,
	SilenceUsage:  true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return readConfig(cmd)
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
	// Set up logger initially, before real config is read
	initLogger()

	defer c.Close()
	err := RootCmd.Execute()
	if rootCtxDone != nil {
		rootCtxDone()
	}
	if err != nil {
		log.Error().Messagef("Failed: %s", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initLogger, initContext)

	RootCmd.SetOut(colorable.NewColorableStdout())
	RootCmd.SetErr(colorable.NewColorableStderr())
	RootCmd.SetUsageTemplate(console.UsageTemplate())

	RootCmd.Flags().BoolVar(&flagLicenseConditions, "license-c", false, "show program's license conditions")
	RootCmd.Flags().BoolVar(&flagLicenseWarranty, "license-w", false, "show program's license warranty")

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", cfgFile, "config file")

	RootCmd.PersistentFlags().Var(&cfg.Client, "client", `Dinkur client: "sqlite" or "grpc"`)
	RootCmd.RegisterFlagCompletionFunc("client", clientComplete)

	RootCmd.PersistentFlags().String("sqlite.path", cfg.Sqlite.Path, "database file")
	RootCmd.PersistentFlags().Bool("sqlite.mkdir", cfg.Sqlite.Mkdir, "create directory for data if it doesn't exist")

	RootCmd.PersistentFlags().String("grpc.address", cfg.GRPC.Address, "address for connecting to Dinkur daemon gRPC API")
	RootCmd.PersistentFlags().String("daemon.address", cfg.Daemon.BindAddress, "bind address for serving Dinkur daemon gRPC API")

	RootCmd.PersistentFlags().Var(&cfg.Log.Level, "log.level", `logging severity: "debug", "info", "warn", "error", or "panic"`)
	RootCmd.RegisterFlagCompletionFunc("log.format", logFormatComplete)

	RootCmd.PersistentFlags().Var(&cfg.Log.Format, "log.format", `logging format: "pretty" or "json"`)
	RootCmd.RegisterFlagCompletionFunc("log.format", logFormatComplete)

	RootCmd.PersistentFlags().Var(&cfg.Log.Color, "log.color", `logging colored output: "auto", "always", or "never"`)
	RootCmd.RegisterFlagCompletionFunc("log.color", logColorComplete)

	RootCmd.PersistentFlags().BoolVarP(&flagVerbose, "verbose", "v", flagVerbose, `enables debug logging (short for --log.level=debug)`)
}

func readConfig(cmd *cobra.Command) error {
	v := viper.New()
	if err := v.BindPFlags(cmd.Root().PersistentFlags()); err != nil {
		return err
	}

	var newCfg *config.Config
	var err error
	if cmd.Flag("config").Changed {
		newCfg, err = config.ReadFile(v, cfgFile)
	} else {
		newCfg, err = config.ReadStandardFiles(v)
	}
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}
	cfg = *newCfg

	// Set up logger again, now that we've read in the new config
	initLogger()

	log.Debug().
		WithString("file", cfg.FileUsed()).
		Message("Loaded configuration.")

	return nil
}

func initLogger() {
	longestLen := logger.LongestScopeNameLength
	logger.ClearOutputs()
	logger.LongestScopeNameLength = longestLen
	level := logger.Level(cfg.Log.Level)
	if flagVerbose {
		level = logger.LevelDebug
	}
	switch cfg.Log.Color {
	case config.LogColorAuto:
		// Do nothing, fatih/color is on auto by default
	case config.LogColorNever:
		color.NoColor = true
	case config.LogColorAlways:
		color.NoColor = false
	}
	if cfg.Log.Format == config.LogFormatPretty {
		prettyConf := consolepretty.DefaultConfig
		prettyConf.DisableDate = true
		prettyConf.DisableCaller = true
		prettyConf.Writer = colorable.NewColorableStderr()
		logger.AddOutput(level, consolepretty.New(prettyConf))
	} else {
		logger.AddOutput(level, consolejson.Default)
	}
}

func initContext() {
	rootCtx, rootCtxDone = signal.NotifyContext(rootCtx, syscall.SIGTERM, syscall.SIGHUP, os.Interrupt, os.Kill)
}

func connectClientOrExit() {
	client, err := connectClient(false)
	if err != nil {
		console.PrintFatal("Error connecting to client:", err)
	}
	c = client
}

func connectClient(skipMigrate bool) (dinkur.Client, error) {
	switch cfg.Client {
	case config.ClientTypeSqlite:
		log.Debug().Message("Using DB client.")
		dbClient, err := connectToDBClient(skipMigrate)
		if err != nil {
			return nil, fmt.Errorf("DB client: %w", err)
		}
		return dbClient, nil
	case config.ClientTypeGRPC:
		log.Debug().Message("Using gRPC client.")
		grpcClient, err := connectToGRPCClient()
		if err != nil {
			return nil, fmt.Errorf("gRPC client: %w", err)
		}
		return grpcClient, nil
	default:
		return nil, fmt.Errorf(`invalid value %q: only "sqlite" or "grpc" may be used`, cfg.Client)
	}
}

func connectToGRPCClient() (dinkur.Client, error) {
	c := dinkurclient.NewClient(cfg.GRPC.Address, dinkurclient.Options{})
	if err := c.Connect(rootCtx); err != nil {
		return nil, err
	}
	if err := c.Ping(rootCtx); err != nil {
		return nil, fmt.Errorf("attempting ping: %w", err)
	}
	checkStatusForAFK(c)
	return c, nil
}

func checkStatusForAFK(c dinkur.Client) {
	status, err := c.GetStatus(rootCtx)
	if err != nil {
		console.PrintFatal("Error getting status:", err)
	}
	if status.AFKSince == nil || status.BackSince == nil {
		return
	}
	activeEntry, err := c.GetActiveEntry(rootCtx)
	if err != nil {
		console.PrintFatal("Error getting active entry:", err)
	}
	if activeEntry == nil {
		return
	}
	promptAFKResolution(c, *activeEntry, *status.AFKSince)
}

func promptAFKResolution(c dinkur.Client, activeEntry dinkur.Entry, afkSince time.Time) {
	res, err := console.PromptAFKResolution(activeEntry, afkSince)
	fmt.Println()
	if err != nil {
		console.PrintFatal("Prompt error:", err)
	}
	if res.Edit != nil {
		update, err := c.UpdateEntry(rootCtx, *res.Edit)
		if err != nil {
			console.PrintFatal("Error editing entry:", err)
		}
		console.PrintEntryEdit(update)
		fmt.Println()
	}
	if res.NewEntry != nil {
		startedEntry, err := c.CreateEntry(rootCtx, *res.NewEntry)
		if err != nil {
			console.PrintFatal("Error starting entry:", err)
		}
		printStartedEntry(startedEntry)
		fmt.Println()
	}
	if err := c.SetStatus(rootCtx, dinkur.EditStatus{}); err != nil {
		console.PrintFatal("Error removing AFK status:", err)
	}
	fmt.Println("Continuing with command...")
	fmt.Println()
}

func connectToDBClient(skipMigrate bool) (dinkur.Client, error) {
	c := dinkurdb.NewClient(cfg.Sqlite.Path, dinkurdb.Options{
		MkdirAll:             cfg.Sqlite.Mkdir,
		DebugLogging:         flagVerbose,
		SkipMigrateOnConnect: skipMigrate,
	})
	return c, c.Connect(rootCtx)
}

func logColorComplete(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return []string{
		"auto\tuse colored terminal output iff session is interactive (default)",
		"always\talways use colored terminal output; may cause issues when piping output",
		"never\tdisables colored terminal output",
	}, cobra.ShellCompDirectiveDefault
}

func logFormatComplete(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return []string{
		"pretty\tprint human-readable log messages (default)",
		"json\tprint machine-readable log messages",
	}, cobra.ShellCompDirectiveDefault
}

func logLevelComplete(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return []string{
		"debug",
		"info\t(default)",
		"warn",
		"error",
		"panic",
	}, cobra.ShellCompDirectiveDefault
}

func clientComplete(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return []string{
		"grpc\tuse grpc client towards a Dinkur daemon",
		"sqlite\tuse database client directly towards an Sqlite3 file (default)",
	}, cobra.ShellCompDirectiveDefault
}

func entryIDComplete(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	client, err := connectClient(true)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	entries, err := client.GetEntryList(rootCtx, dinkur.SearchEntry{
		Limit: 12,
	})
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	completions := make([]string, len(entries))
	var sb strings.Builder
	for i, entry := range entries {
		sb.Reset()
		sb.Grow(len(entry.Name) + 10)
		if entry.End == nil {
			fmt.Fprintf(&sb, "%[1]d\tentry #%[1]d `%[2]s` (active)", entry.ID, entry.Name)
		} else {
			fmt.Fprintf(&sb, "%[1]d\tentry #%[1]d `%[2]s`", entry.ID, entry.Name)
		}
		completions[i] = sb.String()
	}
	return completions, cobra.ShellCompDirectiveDefault
}
