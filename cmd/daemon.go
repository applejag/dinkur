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
	"encoding/json"
	"os"

	"github.com/dinkur/dinkur/internal/console"
	"github.com/dinkur/dinkur/pkg/dinkurd"
	"github.com/dinkur/dinkur/pkg/dinkurdb"
	"github.com/spf13/cobra"
)

// daemonCmd represents the daemon command
var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Starts Dinkur daemon process",
	Long: `The Dinkur daemon hosts a gRPC API on a random port.
This daemon is used by Dinkur clients, and allows more features such as the
"away detection", which is not available when only using the Dinkur CLI.

Information about the daemon, such as which port was selected and what
authentication token can be used, is outputted to the console.`,
	Run: func(cmd *cobra.Command, args []string) {
		dbClient := dinkurdb.NewClient("dinkur.db", dinkurdb.Options{})
		if err := dbClient.Connect(); err != nil {
			console.PrintFatal("Error connecting to database for daemon:", err)
		}
		opt := dinkurd.DefaultOptions
		d := dinkurd.NewDaemon(dbClient, opt)
		defer d.Close()
		enc := json.NewEncoder(os.Stdout)
		enc.Encode(struct {
			Port      uint16  `json:"port"`
			AuthToken *string `json:"authToken"`
		}{
			Port:      opt.Port,
			AuthToken: nil,
		})
		if err := d.Serve(context.TODO()); err != nil {
			console.PrintFatal("Error starting daemon:", err)
		}
	},
}

func init() {
	RootCmd.AddCommand(daemonCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// daemonCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// daemonCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
