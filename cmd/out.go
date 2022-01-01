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
	"os"

	"github.com/dinkur/dinkur/internal/console"
	"github.com/spf13/cobra"
)

// outCmd represents the out command
var outCmd = &cobra.Command{
	Use:     "out",
	Aliases: []string{"o", "end"},
	Short:   "Check out/end the currently active task",
	Long:    ``,
	Run: func(cmd *cobra.Command, args []string) {
		connectClientOrExit()
		stoppedTask, err := c.StopActiveTask(context.Background())
		if err != nil {
			console.PrintFatal("Error stopping task:", err)
		}
		if stoppedTask != nil {
			console.PrintTaskLabel(console.LabelledTask{
				Label: "Stopped task:",
				Task:  *stoppedTask,
			})
		} else {
			fmt.Println("No active task to stop.")
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(outCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// outCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// outCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
