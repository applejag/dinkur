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

	"github.com/dinkur/dinkur/internal/console"
	"github.com/dinkur/dinkur/internal/flagutil"
	"github.com/spf13/cobra"
)

func init() {
	var (
		flagToday bool
		flagWeek  bool
		flagLimit uint = 1000
		flagStart string
		flagEnd   string
	)

	var listCmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "l"},
		Short:   "List your tasks",
		Long:    ``,
		Run: func(cmd *cobra.Command, args []string) {
			console.PrintFatal("Error:", "this feature has not yet been implemented")
			start := flagutil.ParseTime(cmd, "start")
			end := flagutil.ParseTime(cmd, "end")
			fmt.Println("start:", start)
			fmt.Println("end:", end)
		},
	}

	RootCMD.AddCommand(listCmd)

	listCmd.Flags().BoolVar(&flagToday, "today", flagToday, "only list today's tasks")
	listCmd.Flags().BoolVar(&flagWeek, "week", flagWeek, "only list this week's tasks")
	listCmd.Flags().UintVarP(&flagLimit, "limit", "l", flagLimit, "limit the number of results, relative to the last result; 0 will disable limit")
	listCmd.Flags().StringP("start", "s", flagStart, "list tasks starting after or at date time")
	listCmd.Flags().StringP("end", "e", flagEnd, "list tasks ending before or at date time")
}
