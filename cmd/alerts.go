// Dinkur the task time tracking utility.
// <https://github.com/dinkur/dinkur>
//
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
	"time"

	"github.com/dinkur/dinkur/internal/console"
	"github.com/dinkur/dinkur/pkg/dinkur"
	"github.com/spf13/cobra"
)

// alertsCmd represents the test command
var alertsCmd = &cobra.Command{
	Use:   "alerts",
	Short: "Testing alerts streaming",
	Run: func(cmd *cobra.Command, args []string) {
		if flagClient != "grpc" {
			console.PrintFatal("Error running test:", `--client must be set to "grpc"`)
		}
		connectClientOrExit()
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		alertChan, err := c.StreamAlert(ctx)
		if err != nil {
			cancel()
			console.PrintFatal("Error streaming events:", err)
		}
		fmt.Println("Streaming alerts...")
		for {
			alert, ok := <-alertChan
			if !ok {
				cancel()
				fmt.Println("Channel was closed.")
				os.Exit(0)
			}
			fmt.Printf("Received event: #%d %s\n", alert.Alert.ID, alert.Event)
			fmt.Println("  Created at:", alert.Alert.CreatedAt)
			fmt.Println("  Updated at:", alert.Alert.UpdatedAt)
			fmt.Printf("  Type: %T\n", alert.Alert.Type)
			switch alertType := alert.Alert.Type.(type) {
			case dinkur.AlertPlainMessage:
				fmt.Println("  Plain message:")
				fmt.Printf("    Message: %q\n", alertType.Message)
			case dinkur.AlertAFK:
				fmt.Println("  AFK:")
				console.PrintTaskLabel(console.LabelledTask{
					Label: "Active task:",
					Task:  alertType.ActiveTask,
				})
			case dinkur.AlertFormerlyAFK:
				fmt.Println("  Formerly AFK:")
				fmt.Println("    AFK since:", alertType.AFKSince)
				if alertType.ActiveTask != nil {
					console.PrintTaskLabel(console.LabelledTask{
						Label: "    Active task:",
						Task:  *alertType.ActiveTask,
					})
				} else {
					fmt.Println("    Active task: -none-")
				}
			}
			fmt.Println()
		}
	},
}

func init() {
	RootCmd.AddCommand(alertsCmd)
}
