package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of the database and daemon",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("err: this feature has not yet been implemented")
		os.Exit(1)
	},
}

func init() {
	RootCMD.AddCommand(statusCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// statusCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// statusCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
