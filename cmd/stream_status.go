// Dinkur the task time tracking utility.
// <https://github.com/dinkur/dinkur>
//
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
	"time"

	"github.com/dinkur/dinkur/internal/console"
	"github.com/dinkur/dinkur/pkg/config"
	"github.com/spf13/cobra"
)

// streamStatusCmd represents the test command
var streamStatusCmd = &cobra.Command{
	Use:   "status",
	Args:  cobra.NoArgs,
	Short: "Testing status streaming",
	Run: func(cmd *cobra.Command, args []string) {
		if cfg.Client != config.ClientTypeGRPC {
			console.PrintFatal("Error running test:", `--client must be set to "grpc"`)
		}
		connectClientOrExit()
		ctx, cancel := context.WithTimeout(rootCtx, 60*time.Second)
		statusChan, err := c.StreamStatus(ctx)
		if err != nil {
			cancel()
			console.PrintFatal("Error streaming events:", err)
		}
		fmt.Println("Streaming statuses...")
		for {
			ev, ok := <-statusChan
			if !ok {
				cancel()
				fmt.Println("Channel was closed.")
				os.Exit(0)
			}
			logEv := log.Info()
			if ev.Status.AFKSince == nil {
				logEv = logEv.WithString("afkSince", "<null>")
			} else {
				logEv = logEv.WithTime("afkSince", *ev.Status.AFKSince)
			}
			if ev.Status.BackSince == nil {
				logEv = logEv.WithString("backSince", "<null>")
			} else {
				logEv = logEv.WithTime("backSince", *ev.Status.BackSince)
			}

			logEv.Message("Received status.")
			fmt.Println()
		}
	},
}

func init() {
	streamCmd.AddCommand(streamStatusCmd)
}
