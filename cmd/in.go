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
	"strings"
	"time"

	"github.com/dinkur/dinkur/internal/console"
	"github.com/dinkur/dinkur/internal/pflagutil"
	"github.com/dinkur/dinkur/pkg/dinkur"
	"github.com/spf13/cobra"
)

func init() {
	var (
		flagStart     = &pflagutil.Time{Now: true}
		flagEnd       = &pflagutil.Time{}
		flagAfterID   uint
		flagAfterLast bool
		flagBeforeID  uint
	)

	var inCmd = &cobra.Command{
		Use:     "in <entry name>",
		Args:    cobra.ArbitraryArgs,
		Aliases: []string{"i", "start", "new"},
		Short:   "Check in/start tracking a new entry",
		Long:    ``,
		Run: func(cmd *cobra.Command, args []string) {
			connectClientOrExit()
			now := time.Now()
			newEntry := dinkur.NewEntry{
				Name:               strings.Join(args, " "),
				Start:              flagStart.TimePtr(now),
				End:                flagEnd.TimePtr(now),
				StartAfterIDOrZero: flagAfterID,
				EndBeforeIDOrZero:  flagBeforeID,
				StartAfterLast:     flagAfterLast,
			}
			startedEntry, err := c.CreateEntry(rootCtx, newEntry)
			if err != nil {
				console.PrintFatal("Error starting entry:", err)
			}
			printStartedEntry(startedEntry)
		},
	}
	RootCmd.AddCommand(inCmd)

	inCmd.Flags().VarP(flagStart, "start", "s", `start time of entry`)
	inCmd.Flags().VarP(flagEnd, "end", "e", `end time of entry; new entry will not be active if set`)
	inCmd.Flags().UintVarP(&flagAfterID, "after-id", "a", 0, `sets --start time to the end time of entry with ID`)
	inCmd.RegisterFlagCompletionFunc("after-id", entryIDComplete)
	inCmd.Flags().BoolVarP(&flagAfterLast, "after-last", "L", false, `sets --start time to the end time of latest entry`)
	inCmd.Flags().UintVarP(&flagBeforeID, "before-id", "b", 0, `sets --end time to the start time of entry with ID`)
	inCmd.RegisterFlagCompletionFunc("before-id", entryIDComplete)
}

func printStartedEntry(startedEntry dinkur.StartedEntry) {
	var toPrint []console.LabelledEntry
	if startedEntry.Stopped != nil {
		toPrint = append(toPrint, console.LabelledEntry{
			Label: "Stopped entry:",
			Entry: *startedEntry.Stopped,
		})
	}
	noActive := false
	if startedEntry.Started.End != nil {
		toPrint = append(toPrint, console.LabelledEntry{
			Label: "Added entry:",
			Entry: startedEntry.Started,
		})
		noActive = true
	} else {
		toPrint = append(toPrint, console.LabelledEntry{
			Label:      "Started entry:",
			Entry:      startedEntry.Started,
			NoDuration: true,
		})
	}
	console.PrintEntryLabelSlice(toPrint)
	if noActive {
		fmt.Println("You have no active entry.")
	}
}
