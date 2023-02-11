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
	"time"

	"github.com/dinkur/dinkur/internal/console"
	"github.com/dinkur/dinkur/pkg/config"
	"github.com/spf13/cobra"
)

// streamEntriesCmd represents the test command
var streamEntriesCmd = &cobra.Command{
	Use:   "entries",
	Args:  cobra.NoArgs,
	Short: "Testing entry streaming",
	Run: func(cmd *cobra.Command, args []string) {
		if cfg.Client != config.ClientTypeGRPC {
			console.PrintFatal("Error running test:", `--client must be set to "grpc"`)
		}
		connectClientOrExit()
		ctx, cancel := context.WithTimeout(rootCtx, 60*time.Second)
		entryChan, err := c.StreamEntry(ctx)
		if err != nil {
			cancel()
			console.PrintFatal("Error streaming events:", err)
		}
		fmt.Println("Streaming entries...")
		for ev := range entryChan {
			log.Info().
				WithUint("id", ev.Entry.ID).
				WithString("name", ev.Entry.Name).
				WithStringer("event", ev.Event).
				WithTime("createdAt", ev.Entry.CreatedAt).
				WithTime("updatedAt", ev.Entry.UpdatedAt).
				Message("Received entry.")
		}
		cancel()
	},
}

func init() {
	streamCmd.AddCommand(streamEntriesCmd)
}
