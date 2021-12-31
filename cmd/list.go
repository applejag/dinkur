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
	"fmt"
	"os"
	"strings"

	"github.com/dinkur/dinkur/internal/console"
	"github.com/dinkur/dinkur/internal/flagutil"
	"github.com/dinkur/dinkur/pkg/dinkur"
	"github.com/dinkur/dinkur/pkg/timeutil"
	"github.com/spf13/cobra"
)

func init() {
	var (
		flagLimit  uint = 1000
		flagStart  string
		flagEnd    string
		flagOutput = "pretty"
	)

	var listCmd = &cobra.Command{
		Use:     `list [baseline]`,
		Aliases: []string{"ls", "l"},
		Short:   "List your tasks",
		Long: fmt.Sprintf(`Lists all your tasks.

By default, this will only list today's tasks. You can supply an argument
to declare a different baseline. The --start and --end flags will always
take precedence over the baseline.

	%[1]s list all        # list all tasks, i.e. no baseline. Alias: "a"
	%[1]s list past       # list all tasks before now.        Alias: "p"
	%[1]s list future     # list all tasks since now.         Alias: "f"
	%[1]s list today      # (default) list today's tasks.     Alias: "t"
	%[1]s list week       # list this week's tasks.           Alias: "w"
	%[1]s list yesterday  # list yesterday's tasks.           Alias: "y" or "ld"
	%[1]s list lastweek   # list last week's tasks.           Alias: "lw"
	%[1]s list tomorrow   # list tomorrow's tasks.            Alias: "nd"
	%[1]s list nextweek   # list next week's tasks.           Alias: "nw"

Day baselines sets the range 00:00:00 - 24:59:59.
Week baselines sets the range Monday 00:00:00 - Sunday 24:59:59.
`, RootCmd.Name()),
		ValidArgs: []string{
			"all\tlist all tasks",
			"a\talias for 'all'",
			"past\tlist all tasks before now",
			"p\talias for 'past'",
			"future\tlist all tasks since now",
			"f\talias for 'future'",
			"today\tonly list today's tasks (default)",
			"t\talias for 'today'",
			"week\tonly list this week's tasks (monday to sunday)",
			"w\talias for 'week'",
			"yesterday\tonly list yesterday's tasks",
			"y\talias for 'yesterday'",
			"ld\talias for 'yesterday'",
			"lastweek\tonly list last (previous) week's tasks (monday to sunday)",
			"lw\talias for 'lastweek'",
			"tomorrow\tonly list tomorrow's tasks",
			"nd\talias for 'tomorrow'",
			"nextweek\tonly list next week's tasks (monday to sunday)",
			"nw\talias for 'nextweek'",
		},
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			connectClientOrExit()
			search := dinkur.SearchTask{
				Limit:     flagLimit,
				Shorthand: timeutil.TimeSpanThisDay,
			}
			if len(args) > 0 {
				if s, ok := parseShorthand(args[0]); !ok {
					console.PrintFatal("Error parsing argument:", fmt.Sprintf("invalid time span shorthand: %q", args[0]))
				} else {
					search.Shorthand = s
				}
			}
			search.Start = flagutil.ParseTime(cmd, "start")
			search.End = flagutil.ParseTime(cmd, "end")
			printDebugf("--start=%q", search.Start)
			printDebugf("--end=%q", search.End)
			printDebugf("--shorthand=%q", search.Shorthand)
			tasks, err := c.ListTasks(context.Background(), search)
			if err != nil {
				console.PrintFatal("Error getting list of tasks:", err)
			}
			switch strings.ToLower(flagOutput) {
			case "pretty":
				console.PrintTaskList(tasks)
			case "json":
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				enc.Encode(tasks)
			case "jsonl":
				enc := json.NewEncoder(os.Stdout)
				for _, t := range tasks {
					enc.Encode(t)
				}
			default:
				console.PrintFatal("Error parsing --output:", fmt.Errorf("invalid output format: %q", flagOutput))
			}
		},
	}

	RootCmd.AddCommand(listCmd)

	listCmd.Flags().UintVarP(&flagLimit, "limit", "l", flagLimit, "limit the number of results, relative to the last result; 0 will disable limit")
	listCmd.Flags().StringP("start", "s", flagStart, "list tasks starting after or at date time")
	listCmd.Flags().StringP("end", "e", flagEnd, "list tasks ending before or at date time")
	listCmd.Flags().StringVarP(&flagOutput, "output", "o", flagOutput, `set output format: "pretty", "json", or "jsonl"`)
}

func parseShorthand(s string) (timeutil.TimeSpanShorthand, bool) {
	switch strings.ToLower(s) {
	case "all", "a":
		return timeutil.TimeSpanNone, true
	case "past", "p":
		return timeutil.TimeSpanPast, true
	case "future", "f":
		return timeutil.TimeSpanFuture, true
	case "today", "t":
		return timeutil.TimeSpanThisDay, true
	case "week", "w":
		return timeutil.TimeSpanThisWeek, true
	case "yesterday", "y", "ld":
		return timeutil.TimeSpanPrevDay, true
	case "lastweek", "lw":
		return timeutil.TimeSpanPrevWeek, true
	case "tomorrow", "nd":
		return timeutil.TimeSpanNextDay, true
	case "nextweek", "nw":
		return timeutil.TimeSpanNextWeek, true
	default:
		return timeutil.TimeSpanNone, false
	}
}
