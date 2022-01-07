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
	"github.com/spf13/cobra"
)

// tasksCmd represents the test command
var tasksCmd = &cobra.Command{
	Use:   "tasks",
	Args:  cobra.NoArgs,
	Short: "Testing task streaming",
	Run: func(cmd *cobra.Command, args []string) {
		if flagClient != "grpc" {
			console.PrintFatal("Error running test:", `--client must be set to "grpc"`)
		}
		connectClientOrExit()
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		taskChan, err := c.StreamTask(ctx)
		if err != nil {
			cancel()
			console.PrintFatal("Error streaming events:", err)
		}
		fmt.Println("Streaming tasks...")
		for ev := range taskChan {
			log.Info().
				WithUint("id", ev.Task.ID).
				WithString("name", ev.Task.Name).
				WithStringer("event", ev.Event).
				WithTime("createdAt", ev.Task.CreatedAt).
				WithTime("updatedAt", ev.Task.UpdatedAt).
				Message("Received task.")
			fmt.Println()
		}
		cancel()
	},
}

func init() {
	RootCmd.AddCommand(tasksCmd)
}
