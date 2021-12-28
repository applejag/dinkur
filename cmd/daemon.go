package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// daemonCmd represents the daemon command
var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Starts Dinkur daemon process",
	Long: `The Dinkur daemon hosts a gRPC API on a random port.
This daemon is used by Dinkur clients, and allows more features such as the
"away detection", which is not available when only using the Dinkur CLI.

Information about the daemon, such as which port was selected and what
authentication token can be used, is outputted to the console.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("err: this feature has not yet been implemented")
		os.Exit(1)
	},
}

func init() {
	rootCmd.AddCommand(daemonCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// daemonCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// daemonCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
