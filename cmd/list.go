// Dinkur the task time tracking utility.
// <https://github.com/dinkur/dinkur>
//
// Copyright (C) 2021 Kalle Fagerberg
// SPDX-FileCopyrightText: 2021 Kalle Fagerberg
// SPDX-License-Identifier: GPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify it
// under the terms of the GNU General Public License as published by the
// Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for
// more details.
//
// You should have received a copy of the GNU General Public License along
// with this program.  If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dinkur/dinkur/internal/console"
	"github.com/dinkur/dinkur/internal/pflagutil"
	"github.com/dinkur/dinkur/pkg/dinkur"
	"github.com/dinkur/dinkur/pkg/timeutil"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func init() {
	var (
		flagLimit       uint = 1000
		flagStart            = &pflagutil.Time{}
		flagEnd              = &pflagutil.Time{}
		flagRange            = pflagutil.NewTimeRangePtr(timeutil.TimeSpanThisDay)
		flagOutput           = "pretty"
		flagNoHighlight      = false
	)

	var listCmd = &cobra.Command{
		Use:     `list [name search terms]`,
		Args:    cobra.ArbitraryArgs,
		Aliases: []string{"ls", "l"},
		Short:   "List your entries",
		Long: fmt.Sprintf(`Lists all your entries.

Any non-flag arguments are used as entry name search terms. Due to technical
limitation, terms shorter than 3 characters are ignored.

By default, this will only list today's entries. You can supply the --range flag
to declare a different baseline range. The --start and --end flags will always
take precedence over the baseline range.

	%[1]s list --range all        # list all entries, i.e. no baseline.
	%[1]s list --range past       # list all entries before now.
	%[1]s list --range future     # list all entries since now.
	%[1]s list --range today      # (default) list today's entries.
	%[1]s list --range week       # list this week's entries.
	%[1]s list --range yesterday  # list yesterday's entries.
	%[1]s list --range lastweek   # list last week's entries.
	%[1]s list --range tomorrow   # list tomorrow's entries.
	%[1]s list --range nextweek   # list next week's entries.

Day baselines sets the range 00:00:00 - 24:59:59.
Week baselines sets the range Monday 00:00:00 - Sunday 24:59:59.
`, RootCmd.Name()),
		Run: func(cmd *cobra.Command, args []string) {
			connectClientOrExit()
			rand.Seed(time.Now().UnixMicro())
			now := time.Now()
			search := dinkur.SearchEntry{
				Limit:     flagLimit,
				Start:     flagStart.TimePtr(now),
				End:       flagEnd.TimePtr(now),
				Shorthand: flagRange.TimeSpanShorthand(),
				NameFuzzy: strings.Join(args, " "),
			}
			if strings.EqualFold(flagOutput, "pretty") && !flagNoHighlight {
				search.NameHighlightStart = fmt.Sprintf(">!@%d#>", rand.Intn(255))
				search.NameHighlightEnd = fmt.Sprintf("<!@%d#<", rand.Intn(255))
			}
			log.Debug().
				WithStringf("--start", "%v", search.Start).
				WithStringf("--end", "%v", search.End).
				WithStringf("--shorthand", "%v", search.Shorthand).
				Message("Flags")
			entries, err := c.GetEntryList(rootCtx, search)
			if err != nil {
				console.PrintFatal("Error getting list of entries:", err)
			}
			switch strings.ToLower(flagOutput) {
			case "pretty":
				searchStart, searchEnd := search.NameHighlightStart, search.NameHighlightEnd
				if len(args) == 0 {
					searchStart, searchEnd = "", ""
				}
				console.PrintEntryListSearched(entries, searchStart, searchEnd)
			case "pretty-compact", "pc":
				console.PrintEntryListCompact(entries)
			case "json":
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				if err := enc.Encode(entries); err != nil {
					console.PrintFatal("Error encoding entries as JSON:", err)
				}
			case "json-line":
				enc := json.NewEncoder(os.Stdout)
				for _, t := range entries {
					if err := enc.Encode(t); err != nil {
						console.PrintFatal(fmt.Sprintf("Error encoding entry #%d as JSON:", t.ID), err)
					}
				}
			case "yaml":
				enc := yaml.NewEncoder(os.Stdout)
				enc.SetIndent(2)
				if err := enc.Encode(entries); err != nil {
					console.PrintFatal("Error encoding entries as YAML:", err)
				}
			case "xml":
				enc := xml.NewEncoder(os.Stdout)
				enc.Indent("", "    ")
				if err := enc.Encode(entries); err != nil {
					console.PrintFatal("Error encoding entries as XML:", err)
				}
				fmt.Println()
			case "xml-line":
				enc := xml.NewEncoder(os.Stdout)
				for _, t := range entries {
					if err := enc.Encode(t); err != nil {
						fmt.Println()
						console.PrintFatal(fmt.Sprintf("Error encoding entry #%d as XML:", t.ID), err)
					}
					fmt.Println()
				}
			case "csv":
				w := csv.NewWriter(os.Stdout)
				var records [][]string
				for _, t := range entries {
					records = append(records, convEntryCSVRecord(t))
				}
				if err := w.WriteAll(records); err != nil {
					console.PrintFatal("Error encoding entries as CSV:", err)
				}
			case "csv-header":
				w := csv.NewWriter(os.Stdout)
				records := [][]string{entryCSVHeaderRecord()}
				for _, t := range entries {
					records = append(records, convEntryCSVRecord(t))
				}
				if err := w.WriteAll(records); err != nil {
					console.PrintFatal("Error encoding entries as CSV:", err)
				}
			default:
				console.PrintFatal("Error parsing --output:", fmt.Errorf("invalid output format: %q", flagOutput))
			}
		},
	}

	RootCmd.AddCommand(listCmd)

	listCmd.Flags().UintVarP(&flagLimit, "limit", "l", flagLimit, "limit the number of results, relative to the last result; 0 will disable limit")
	listCmd.Flags().VarP(flagStart, "start", "s", "list entries starting after or at date time")
	listCmd.Flags().VarP(flagEnd, "end", "e", "list entries ending before or at date time")
	listCmd.Flags().VarP(flagRange, "range", "r", "baseline time range")
	listCmd.RegisterFlagCompletionFunc("range", pflagutil.TimeRangeCompletion)
	listCmd.Flags().StringVarP(&flagOutput, "output", "o", flagOutput, `set output format: "pretty", "pretty-compact [pc]", "json", "json-line", "yaml", "xml", "xml-line", "csv", "csv-header"`)
	listCmd.RegisterFlagCompletionFunc("output", outputFormatComplete)
	listCmd.Flags().BoolVar(&flagNoHighlight, "no-highlight", false, `disables search highlighting in "pretty" output`)
}

func outputFormatComplete(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return []string{
		"pretty\thuman readable and colored table formatting (default)",
		"json\ta single indented JSON array containing all entries",
		"json-line\teach entry JSON object on a separate line",
		"yaml\tYAML array of entries",
		"xml\tXML list of entries",
		"xml-line\teach entry XML element on a separate line",
		"csv\teach entry on a separate line with field as comma-separated-values",
		"csv-header\tsame as --output=csv, but with additional header row",
	}, cobra.ShellCompDirectiveDefault
}

func entryCSVHeaderRecord() []string {
	return []string{
		"ID",
		"Created at",
		"Updated at",
		"Name",
		"Start",
		"End",
	}
}

func convEntryCSVRecord(entry dinkur.Entry) []string {
	const timeLayout = time.RFC3339Nano
	endStr := ""
	if entry.End != nil {
		endStr = entry.End.String()
	}
	return []string{
		strconv.FormatUint(uint64(entry.ID), 10),
		entry.CreatedAt.Format(timeLayout),
		entry.UpdatedAt.Format(timeLayout),
		entry.Name,
		entry.Start.Format(timeLayout),
		endStr,
	}
}
