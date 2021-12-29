package flagutil

import (
	"fmt"
	"time"

	"github.com/dinkur/dinkur/internal/console"
	"github.com/dinkur/dinkur/internal/fuzzytime"
	"github.com/spf13/cobra"
)

func ParseTime(cmd *cobra.Command, name string) *time.Time {
	f := cmd.Flags().Lookup(name)
	if f == nil || !f.Changed {
		return nil
	}
	val := f.Value.String()
	if val == "" {
		printFatal(name, "cannot be empty")
	}
	start, err := fuzzytime.Parse(val)
	if err != nil {
		printFatal(name, err)
	}
	return &start
}

func printFatal(name string, v interface{}) {
	console.PrintFatal(fmt.Sprintf("Error parsing --%s:", name), v)
}
