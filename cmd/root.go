package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/dinkur/dinkur/internal/cfgpath"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile = cfgpath.Path()

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "dinkur",
	Short: "The Dinkur CLI",
	Long:  `Through these subcommands you can access your time-tracked tasks.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", cfgFile, "config file")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetConfigFile(cfgFile)

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	} else if !errors.As(err, &viper.ConfigFileNotFoundError{}) && !errors.Is(err, os.ErrNotExist) {
		fmt.Fprintf(os.Stderr, "Error reading config: (%[1]T) %[1]v\n", err)
		os.Exit(1)
	}
}
