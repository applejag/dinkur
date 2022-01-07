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
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/dinkur/dinkur/internal/console"
	"github.com/dinkur/dinkur/internal/pflagutil"
	"github.com/dinkur/dinkur/pkg/dinkur"
	"github.com/dinkur/dinkur/pkg/timeutil"
	"github.com/spf13/cobra"
)

func init() {
	var (
		flagLimit       uint = 1000
		flagStart            = &pflagutil.Time{}
		flagEnd              = &pflagutil.Time{}
		flagRange            = pflagutil.NewTimeRangePtr(timeutil.TimeSpanThisDay)
		flagOutput           = "pretty"
		flagNoHighlight      = false
	)

	var listCmd = &cobra.Command{
		Use:     `list [name search terms]`,
		Args:    cobra.ArbitraryArgs,
		Aliases: []string{"ls", "l"},
		Short:   "List your tasks",
		Long: fmt.Sprintf(`Lists all your tasks.

Any non-flag arguments are used as task name search terms. Due to technical
limitation, terms shorter than 3 characters are ignored.

By default, this will only list today's tasks. You can supply the --range flag
to declare a different baseline range. The --start and --end flags will always
take precedence over the baseline range.

	%[1]s list --range all        # list all tasks, i.e. no baseline.
	%[1]s list --range past       # list all tasks before now.
	%[1]s list --range future     # list all tasks since now.
	%[1]s list --range today      # (default) list today's tasks.
	%[1]s list --range week       # list this week's tasks.
	%[1]s list --range yesterday  # list yesterday's tasks.
	%[1]s list --range lastweek   # list last week's tasks.
	%[1]s list --range tomorrow   # list tomorrow's tasks.
	%[1]s list --range nextweek   # list next week's tasks.

Day baselines sets the range 00:00:00 - 24:59:59.
Week baselines sets the range Monday 00:00:00 - Sunday 24:59:59.
`, RootCmd.Name()),
		Run: func(cmd *cobra.Command, args []string) {
			connectClientOrExit()
			activeTask, err := c.ActiveTask(context.Background())
			if err == nil && activeTask != nil {
				res, err := console.PromptAFKResolution(dinkur.AlertFormerlyAFK{
					AFKSince:   time.Now().Add(-15 * time.Minute),
					ActiveTask: *activeTask,
				})
				if err != nil {
					console.PrintFatal("Prompt error:", err)
				}
				fmt.Println()
				fmt.Println("Resolution:")
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				enc.Encode(res)
				os.Exit(1)
			}
			rand.Seed(time.Now().UnixMicro())
			search := dinkur.SearchTask{
				Limit:     flagLimit,
				Start:     flagStart.TimePtr(),
				End:       flagEnd.TimePtr(),
				Shorthand: flagRange.TimeSpanShorthand(),
				NameFuzzy: strings.Join(args, " "),
			}
			if strings.EqualFold(flagOutput, "pretty") && !flagNoHighlight {
				search.NameHighlightStart = fmt.Sprintf(">!@%d#>", rand.Intn(255))
				search.NameHighlightEnd = fmt.Sprintf("<!@%d#<", rand.Intn(255))
			}
			log.Debug().
				WithStringf("--start", "%v", search.Start).
				WithStringf("--end", "%v", search.End).
				WithStringf("--shorthand", "%v", search.Shorthand).
				Message("Flags")
			tasks, err := c.ListTasks(context.Background(), search)
			if err != nil {
				console.PrintFatal("Error getting list of tasks:", err)
			}
			switch strings.ToLower(flagOutput) {
			case "pretty":
				searchStart, searchEnd := search.NameHighlightStart, search.NameHighlightEnd
				if len(args) == 0 {
					searchStart, searchEnd = "", ""
				}
				console.PrintTaskListSearched(tasks, searchStart, searchEnd)
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
	listCmd.Flags().VarP(flagStart, "start", "s", "list tasks starting after or at date time")
	listCmd.Flags().VarP(flagEnd, "end", "e", "list tasks ending before or at date time")
	listCmd.Flags().VarP(flagRange, "range", "r", "baseline time range")
	listCmd.RegisterFlagCompletionFunc("range", pflagutil.TimeRangeCompletion)
	listCmd.Flags().StringVarP(&flagOutput, "output", "o", flagOutput, `set output format: "pretty", "json", or "jsonl"`)
	listCmd.RegisterFlagCompletionFunc("output", outputFormatComplete)
	listCmd.Flags().BoolVar(&flagNoHighlight, "no-highlight", false, `disables search highlighting in "pretty" output`)
}

func outputFormatComplete(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return []string{
		"pretty\thuman readable and colored table formatting (default)",
		"json\ta single indented JSON array containing all tasks",
		"jsonl\teach task JSON object on a separate line",
	}, cobra.ShellCompDirectiveDefault
}
