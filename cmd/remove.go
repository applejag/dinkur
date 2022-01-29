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
	"fmt"
	"time"

	"github.com/dinkur/dinkur/internal/console"
	"github.com/spf13/cobra"
)

func init() {
	var (
		flagID  uint
		flagYes bool
	)

	// removeCmd represents the remove command
	var removeCmd = &cobra.Command{
		Use:     "remove",
		Args:    cobra.NoArgs,
		Aliases: []string{"rm", "r"},
		Short:   "Removes a entry",
		Long: `Removes a entry from your entry data store.
You must provide the flag --id to specify which entry to remove.
No bulk removal is supported.

Warning: Removing a entry cannot be undone!`,
		Run: func(cmd *cobra.Command, args []string) {
			connectClientOrExit()
			if !flagYes {
				entry, err := c.GetEntry(rootCtx, flagID)
				if err != nil {
					console.PrintFatal("Error getting entry:", err)
				}
				err = console.PromptEntryRemoval(entry)
				if err != nil {
					console.PrintFatal("Prompt error:", err)
				}
			}
			removedEntry, err := c.DeleteEntry(rootCtx, flagID)
			if err != nil {
				console.PrintFatal("Error removing entry:", err)
			}
			console.PrintEntryLabel(console.LabelledEntry{
				Label: "Deleted entry:",
				Entry: removedEntry,
			})
			fmt.Println()
			fmt.Println("If this was a mistake, you can add it back in with:")
			if removedEntry.End != nil {
				fmt.Printf("  $ dinkur in --start %q --end %q %q\n",
					removedEntry.Start.Format(time.RFC3339),
					removedEntry.End.Format(time.RFC3339),
					removedEntry.Name)
			} else {
				fmt.Printf("  $ dinkur in --start %q %q\n",
					removedEntry.Start.Format(time.RFC3339),
					removedEntry.Name)
			}
		},
	}

	RootCmd.AddCommand(removeCmd)

	removeCmd.Flags().UintVarP(&flagID, "id", "i", 0, "ID of entry to be removed (required)")
	removeCmd.MarkFlagRequired("id")
	removeCmd.RegisterFlagCompletionFunc("id", entryIDComplete)
	removeCmd.Flags().BoolVarP(&flagYes, "yes", "y", false, "skip confirmation prompt")
}
