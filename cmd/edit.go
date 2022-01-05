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
	"strings"

	"github.com/dinkur/dinkur/internal/console"
	"github.com/dinkur/dinkur/internal/pflagutil"
	"github.com/dinkur/dinkur/pkg/dinkur"
	"github.com/spf13/cobra"
)

func init() {
	var (
		flagID     uint
		flagAppend bool
		flagStart  = &pflagutil.Time{}
		flagEnd    = &pflagutil.Time{}
	)

	var editCmd = &cobra.Command{
		Use:     "edit [new name of task]",
		Args:    cobra.ArbitraryArgs,
		Aliases: []string{"e"},
		Short:   "Edit the latest or a specific task",
		Long: `Applies changes to the currently active task, or the latest task, or
a specific task using the --id or -i flag.`,
		Run: func(cmd *cobra.Command, args []string) {
			connectClientOrExit()
			log.Debug().
				WithStringer("start", flagStart).
				WithStringer("end", flagEnd).
				WithBool("append", flagAppend).
				Message("Flags")
			edit := dinkur.EditTask{
				IDOrZero:   flagID,
				Start:      flagStart.TimePtr(),
				End:        flagEnd.TimePtr(),
				AppendName: flagAppend,
			}
			if len(args) > 0 {
				name := strings.Join(args, " ")
				edit.Name = &name
			}
			update, err := c.EditTask(context.Background(), edit)
			if err != nil {
				console.PrintFatal("Error editing task:", err)
			}
			console.PrintTaskEdit(update)
		},
	}

	RootCmd.AddCommand(editCmd)

	editCmd.Flags().VarP(flagStart, "start", "s", `start time of task`)
	editCmd.Flags().VarP(flagEnd, "end", "e", `end time of task; task will be unmarked as active if set`)
	editCmd.Flags().BoolVarP(&flagAppend, "append", "z", flagAppend, `add name to the end of the existing name, instead of replacing it`)
	editCmd.Flags().UintVarP(&flagID, "id", "i", 0, `ID of task (default is active or latest task)`)
}
