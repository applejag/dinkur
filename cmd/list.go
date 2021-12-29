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
	"fmt"
	"strings"

	"github.com/dinkur/dinkur/internal/console"
	"github.com/dinkur/dinkur/internal/flagutil"
	"github.com/dinkur/dinkur/pkg/dinkurdb"
	"github.com/dinkur/dinkur/pkg/timeutil"
	"github.com/spf13/cobra"
)

func init() {
	var (
		flagLimit uint = 1000
		flagStart string
		flagEnd   string
	)

	var listCmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "l"},
		Short:   "List your tasks",
		Long:    ``,
		ValidArgs: []string{
			"today\tonly list today's tasks (default)",
			"t\talias for 'today'",
			"week\tonly list this week's tasks (monday to sunday)",
			"w\talias for 'week'",
		},
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			connectAndMigrateDB()
			search := dinkurdb.SearchTask{
				Limit:     flagLimit,
				Shorthand: timeutil.TimeSpanThisDay,
			}
			if len(args) > 0 {
				search.Shorthand = parseShorthand(args[0])
				if search.Shorthand == timeutil.TimeSpanNone {
					console.PrintFatal("Error parsing argument:", fmt.Sprintf("invalid time span shorthand: %q", args[0]))
				}
			}
			search.Start = flagutil.ParseTime(cmd, "start")
			search.End = flagutil.ParseTime(cmd, "end")
			tasks, err := db.ListTasks(search)
			if err != nil {
				console.PrintFatal("Error getting list of tasks:", err)
			}
			for _, task := range tasks {
				console.PrintTaskWithDuration("-", task)
			}
			fmt.Printf("Total: %d tasks\n", len(tasks))
		},
	}

	RootCMD.AddCommand(listCmd)

	listCmd.Flags().UintVarP(&flagLimit, "limit", "l", flagLimit, "limit the number of results, relative to the last result; 0 will disable limit")
	listCmd.Flags().StringP("start", "s", flagStart, "list tasks starting after or at date time")
	listCmd.Flags().StringP("end", "e", flagEnd, "list tasks ending before or at date time")
}

func parseShorthand(s string) timeutil.TimeSpanShorthand {
	switch strings.ToLower(s) {
	case "today", "t":
		return timeutil.TimeSpanThisDay
	case "week", "w":
		return timeutil.TimeSpanThisWeek
	default:
		return timeutil.TimeSpanNone
	}
}
