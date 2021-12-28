package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// inCmd represents the in command
var inCmd = &cobra.Command{
	Use:     "in",
	Aliases: []string{"i", "start", "new"},
	Short:   "Check in/start tracking a new task",
	Long:    ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("err: this feature has not yet been implemented")
		os.Exit(1)
	},
}

func init() {
	rootCmd.AddCommand(inCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// inCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// inCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
