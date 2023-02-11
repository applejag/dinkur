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
	"os"
	"os/signal"
	"syscall"

	"github.com/dinkur/dinkur/internal/console"
	"github.com/dinkur/dinkur/pkg/dinkurd"
	"github.com/spf13/cobra"
)

// daemonCmd represents the daemon command
var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Args:  cobra.NoArgs,
	Short: "Starts Dinkur daemon process",
	Long: `The Dinkur daemon hosts a gRPC API on a random port.
This daemon is used by Dinkur clients, and allows more features such as the
"away detection", which is not available when only using the Dinkur CLI.

Information about the daemon, such as which port was selected and what
authentication token can be used, is outputted to the console.`,
	Run: func(cmd *cobra.Command, args []string) {
		dbClient, err := connectToDBClient(false)
		if err != nil {
			console.PrintFatal("Error connecting to database for daemon:", err)
		}
		opt := dinkurd.DefaultOptions
		opt.BindAddress = cfg.Daemon.BindAddress
		d := dinkurd.NewDaemon(dbClient, opt)
		defer d.Close()
		if err := d.Serve(contextWithOSInterrupt(rootCtx)); err != nil {
			console.PrintFatal("Error starting daemon:", err)
		}
	},
}

func init() {
	RootCmd.AddCommand(daemonCmd)
}

func contextWithOSInterrupt(ctx context.Context) context.Context {
	c := make(chan os.Signal)
	newCtx, done := context.WithCancel(ctx)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case <-c:
			log.Debug().Message("Detected Ctrl+C, attempting graceful hault.")
			done()
			close(c)
		case <-ctx.Done():
			close(c)
		}
	}()
	return newCtx
}
