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
	"strings"

	"github.com/dinkur/dinkur/internal/console"
	"github.com/dinkur/dinkur/internal/pflagutil"
	"github.com/dinkur/dinkur/pkg/dinkur"
	"github.com/spf13/cobra"
)

func init() {
	var (
		flagStart = &pflagutil.Time{Now: true}
		flagEnd   = &pflagutil.Time{}
	)

	var inCmd = &cobra.Command{
		Use:     "in <task name>",
		Args:    cobra.ArbitraryArgs,
		Aliases: []string{"i", "start", "new"},
		Short:   "Check in/start tracking a new task",
		Long:    ``,
		Run: func(cmd *cobra.Command, args []string) {
			connectClientOrExit()
			newTask := dinkur.NewTask{
				Name:  strings.Join(args, " "),
				Start: flagStart.TimePtr(),
				End:   flagEnd.TimePtr(),
			}
			startedTask, err := c.StartTask(context.Background(), newTask)
			var toPrint []console.LabelledTask
			if err != nil {
				console.PrintFatal("Error starting task:", err)
			}
			if startedTask.Previous != nil {
				toPrint = append(toPrint, console.LabelledTask{
					Label: "Stopped task:",
					Task:  *startedTask.Previous,
				})
			}
			noActive := false
			if startedTask.New.End != nil {
				toPrint = append(toPrint, console.LabelledTask{
					Label: "Added task:",
					Task:  startedTask.New,
				})
				noActive = true
			} else {
				toPrint = append(toPrint, console.LabelledTask{
					Label:      "Started task:",
					Task:       startedTask.New,
					NoDuration: true,
				})
			}
			console.PrintTaskLabelSlice(toPrint)
			if noActive {
				fmt.Println("You have no active task.")
			}
		},
	}
	RootCmd.AddCommand(inCmd)

	inCmd.Flags().VarP(flagStart, "start", "s", `start time of task`)
	inCmd.Flags().VarP(flagEnd, "end", "e", `end time of task; new task will not be active if set`)
}
