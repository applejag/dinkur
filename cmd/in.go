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
	"os"
	"strings"

	"github.com/dinkur/dinkur/internal/console"
	"github.com/dinkur/dinkur/internal/fuzzytime"
	"github.com/dinkur/dinkur/pkg/dinkurdb"
	"github.com/spf13/cobra"
)

func init() {
	var (
		flagStart string
		flagEnd   string
	)

	var inCmd = &cobra.Command{
		Use:     "in",
		Args:    cobra.ArbitraryArgs,
		Aliases: []string{"i", "start", "new"},
		Short:   "Check in/start tracking a new task",
		Long:    ``,
		Run: func(cmd *cobra.Command, args []string) {
			connectAndMigrateDB()
			newTask := dinkurdb.NewTask{Name: strings.Join(args, " ")}
			if cmd.Flags().Changed("start") {
				if flagStart == "" {
					fmt.Fprintln(os.Stderr, "Error parsing --start: cannot be empty")
					os.Exit(1)
				}
				start, err := fuzzytime.Parse(flagStart)
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error parsing --start:", err)
					os.Exit(1)
				}
				newTask.Start = &start
			}
			if cmd.Flags().Changed("end") {
				if flagEnd == "" {
					fmt.Fprintln(os.Stderr, "Error parsing --end: cannot be empty")
					os.Exit(1)
				}
				end, err := fuzzytime.Parse(flagEnd)
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error parsing --end:", err)
					os.Exit(1)
				}
				newTask.End = &end
			}
			startedTask, err := db.StartTask(newTask)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error starting task:", err)
				os.Exit(1)
			}
			if startedTask.Previous != nil {
				console.PrintTaskWithDuration("Stopped task:", *startedTask.Previous)
			}
			if startedTask.New.End != nil {
				console.PrintTaskWithDuration("Started task:", startedTask.New)
				fmt.Println("You have no active task.")
			} else {
				console.PrintTask("Started task:", startedTask.New)
			}
		},
	}
	RootCMD.AddCommand(inCmd)

	inCmd.Flags().StringVarP(&flagStart, "start", "s", "now", `start time of task`)
	inCmd.Flags().StringVarP(&flagEnd, "end", "e", "", `end time of task; new task will not be active if set`)
}
