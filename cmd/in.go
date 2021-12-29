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

	"github.com/dinkur/dinkur/internal/fuzzytime"
	"github.com/dinkur/dinkur/pkg/dinkurdb"
	"github.com/spf13/cobra"
)

func init() {
	var (
		flagStart string
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
			if flagStart != "" {
				start, err := fuzzytime.Parse(flagStart)
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error parsing --start:", err)
					os.Exit(1)
				}
				newTask.Start = &start
			}
			startedTask, err := db.StartTask(newTask)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error starting task:", err)
				os.Exit(1)
			}
			if startedTask.Previous != nil {
				fmt.Println("Stopped task:", *startedTask.Previous)
			}
			fmt.Println("Started task:", startedTask.New)
		},
	}
	RootCMD.AddCommand(inCmd)

	inCmd.Flags().StringVarP(&flagStart, "start", "s", "now", `start time of task`)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// inCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// inCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
