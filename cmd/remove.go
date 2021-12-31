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
	"fmt"
	"time"

	"github.com/dinkur/dinkur/internal/console"
	"github.com/spf13/cobra"
)

func init() {
	var (
		flagID uint
	)

	// removeCmd represents the remove command
	var removeCmd = &cobra.Command{
		Use:     "remove",
		Aliases: []string{"rm", "r"},
		Short:   "Removes a task",
		Long: `Removes a task from your task data store.
You must provide the flag --id to specify which task to remove.
No bulk removal is supported.

Warning: Removing a task cannot be undone!`,
		Run: func(cmd *cobra.Command, args []string) {
			connectClientOrExit()
			task, err := c.DeleteTask(context.Background(), flagID)
			if err != nil {
				console.PrintFatal("Error removing task:", err)
			}
			console.PrintTaskWithDuration("Deleted task:", task)
			fmt.Println()
			fmt.Println("If this was a mistake, you can add it back in with:")
			if task.End != nil {
				fmt.Printf("  $ dinkur in --start %q --end %q %q\n",
					task.Start.Format(time.RFC3339),
					task.End.Format(time.RFC3339),
					task.Name)
			} else {
				fmt.Printf("  $ dinkur in --start %q %q\n",
					task.Start.Format(time.RFC3339),
					task.Name)
			}
		},
	}

	RootCmd.AddCommand(removeCmd)

	removeCmd.Flags().UintVarP(&flagID, "id", "i", 0, "ID of task to be removed (required)")
	removeCmd.MarkFlagRequired("id")
}
