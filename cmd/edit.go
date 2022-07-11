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
	"strings"

	"github.com/dinkur/dinkur/internal/console"
	"github.com/dinkur/dinkur/internal/pflagutil"
	"github.com/dinkur/dinkur/pkg/dinkur"
	"github.com/spf13/cobra"
)

func init() {
	var (
		flagID        uint
		flagAppend    bool
		flagStart     = &pflagutil.Time{}
		flagEnd       = &pflagutil.Time{}
		flagAfterID   uint
		flagAfterLast bool
		flagBeforeID  uint
	)

	var editCmd = &cobra.Command{
		Use:     "edit [new name of entry]",
		Args:    cobra.ArbitraryArgs,
		Aliases: []string{"e"},
		Short:   "Edit the latest or a specific entry",
		Long: `Applies changes to the currently active entry, or the latest entry, or
a specific entry using the --id or -i flag.`,
		Run: func(cmd *cobra.Command, args []string) {
			connectClientOrExit()
			log.Debug().
				WithStringer("start", flagStart).
				WithStringer("end", flagEnd).
				WithBool("append", flagAppend).
				Message("Flags")
			edit := dinkur.EditEntry{
				IDOrZero:           flagID,
				Start:              flagStart.TimePtr(),
				End:                flagEnd.TimePtr(),
				AppendName:         flagAppend,
				StartAfterIDOrZero: flagAfterID,
				EndBeforeIDOrZero:  flagBeforeID,
				StartAfterLast:     flagAfterLast,
			}
			if len(args) > 0 {
				name := strings.Join(args, " ")
				edit.Name = &name
			}
			update, err := c.UpdateEntry(rootCtx, edit)
			if err != nil {
				console.PrintFatal("Error editing entry:", err)
			}
			console.PrintEntryEdit(update)
		},
	}

	RootCmd.AddCommand(editCmd)

	editCmd.Flags().VarP(flagStart, "start", "s", `start time of entry`)
	editCmd.Flags().VarP(flagEnd, "end", "e", `end time of entry; entry will be unmarked as active if set`)
	editCmd.Flags().BoolVarP(&flagAppend, "append", "z", flagAppend, `add name to the end of the existing name, instead of replacing it`)
	editCmd.Flags().UintVarP(&flagID, "id", "i", 0, `ID of entry (default is active or latest entry)`)
	editCmd.RegisterFlagCompletionFunc("id", entryIDComplete)
	editCmd.Flags().UintVarP(&flagAfterID, "after-id", "a", 0, `sets --start time to the end time of entry with ID`)
	editCmd.RegisterFlagCompletionFunc("after-id", entryIDComplete)
	editCmd.Flags().BoolVarP(&flagAfterLast, "after-last", "L", false, `sets --start time to the end time of latest entry`)
	editCmd.Flags().UintVarP(&flagBeforeID, "before-id", "b", 0, `sets --end time to the start time of entry with ID`)
	editCmd.RegisterFlagCompletionFunc("before-id", entryIDComplete)
}
