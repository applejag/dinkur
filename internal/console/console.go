// Dinkur the task time tracking utility.
// <https://github.com/dinkur/dinkur>
//
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

// Package console contains code to pretty-print different types to the console.
package console

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/dinkur/dinkur/pkg/dinkur"
	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
)

var (
	stdout          = colorable.NewColorableStdout()
	stderr          = colorable.NewColorableStderr()
	timeFormatLong  = "Jan 02 15:04"
	timeFormatShort = "15:04"

	entryIDColor              = color.New(color.FgHiBlack)
	entryLabelColor           = color.New(color.FgWhite, color.Italic)
	entryNameColor            = color.New(color.FgYellow)
	entryNameHighlightColor   = color.New(color.FgHiYellow, color.Underline)
	entryNameHighlightReplace = entryNameHighlightColor.Sprint("$1")
	entryNameQuote            = "`"
	entryNameFormat           = entryNameQuote + "%s" + entryNameQuote
	entryTimeDelimColor       = color.New(color.FgHiBlack)
	entryMonthColor           = color.New(color.FgGreen)
	entryWeekColor            = color.New(color.FgGreen)
	entryDayColor             = color.New(color.FgGreen)
	entryStartColor           = color.New(color.FgGreen)
	entryEndColor             = color.New(color.FgGreen)
	entryEndNilColor          = color.New(color.FgHiBlack, color.Italic)
	entryEndNilTextNow        = "now…"
	entryEndNilTextActive     = "active…"
	entryDurationColor        = color.New(color.FgCyan)
	entryEditDelimColor       = color.New(color.FgHiMagenta)
	entryEditNoneColor        = color.New(color.FgHiBlack, color.Italic)

	entryEditPrefix   = "  "
	entryEditNoChange = "No changes were applied."
	entryEditSpacing  = "   "
	entryEditDelim    = "=>"

	fatalLabelColor = color.New(color.FgHiRed, color.Bold)
	fatalValueColor = color.New(color.FgRed)

	tableEmptyColor        = color.New(color.FgHiBlack, color.Italic)
	tableEmptyText         = "No results to display."
	tableHeaderColor       = color.New(color.FgHiBlack)
	tableSummaryColor      = color.New(color.FgHiBlack, color.Italic)
	tableWeekSummaryColor  = color.New(color.FgHiBlack)
	tableMonthSummaryColor = color.New(color.FgHiBlack)
	tableCellEmptyText     = "-"
	tableCellEmptyColor    = color.New(color.FgHiBlack)

	usageHeaderColor = color.New(color.FgYellow, color.Underline, color.Italic)
	usageHelpColor   = color.New(color.FgHiBlack, color.Italic)

	promptWarnIconColor  = color.New(color.FgHiRed, color.Bold)
	promptWarnIconText   = "!"
	promptErrorColor     = color.New(color.FgRed)
	promptCtrlCHelpColor = color.New(color.FgHiBlack, color.Italic)
)

// LabelledEntry holds a string label and a entry. Used when printing multiple
// labelled entries together.
type LabelledEntry struct {
	Label      string
	Entry      dinkur.Entry
	NoDuration bool
}

// PrintEntryLabel writes a label string followed by a formatted entry to STDOUT.
func PrintEntryLabel(labelled LabelledEntry) {
	var t table
	t.SetSpacing("  ")
	t.WriteColoredRow(tableHeaderColor, "", "ID", "NAME", "START", "END", "DURATION")
	writeCellsLabelledEntry(&t, labelled)
	t.CommitRow()
	t.Fprintln(stdout)
}

// PrintEntryLabelSlice writes a table of label strings followed by a formatted
// entry to STDOUT.
func PrintEntryLabelSlice(slice []LabelledEntry) {
	var t table
	t.SetSpacing("  ")
	t.WriteColoredRow(tableHeaderColor, "", "ID", "NAME", "START", "END", "DURATION")
	for _, lbl := range slice {
		writeCellsLabelledEntry(&t, lbl)
		t.CommitRow()
	}
	t.Fprintln(stdout)
}

// PrintFatal writes a label and some error value to STDERR and then exits the
// application with status code 1.
func PrintFatal(label string, v any) {
	var sb strings.Builder
	fatalLabelColor.Fprint(&sb, label)
	sb.WriteByte(' ')
	fatalValueColor.Fprint(&sb, v)
	fmt.Fprintln(stderr, sb.String())
	os.Exit(1)
}

// PrintEntryEdit writes a formatted entry and highlights any edits made to it,
// by diffing the before and after entries, to STDOUT.
func PrintEntryEdit(update dinkur.UpdatedEntry) {
	var sb strings.Builder
	entryLabelColor.Fprint(&sb, "Updated entry ")
	entryIDColor.Fprint(&sb, "#", update.After.ID)
	sb.WriteByte(' ')
	writeEntryName(&sb, update.After.Name)
	entryLabelColor.Fprint(&sb, ":")
	fmt.Fprintln(stdout, sb.String())

	var t table
	t.SetPrefix(entryEditPrefix)
	t.SetSpacing(entryEditSpacing)
	if update.Before.Name != update.After.Name {
		writeCellEntryName(&t, update.Before.Name)
		t.WriteCellColor(entryEditDelim, entryEditDelimColor)
		writeCellEntryName(&t, update.After.Name)
		t.CommitRow()
	}
	if !timesEqual(update.Before.Start, update.After.Start) ||
		!timesPtrsEqual(update.Before.End, update.After.End) {
		writeCellEntryTimeSpanDuration(&t, update.Before.Start, update.Before.End, update.Before.Elapsed())
		t.WriteCellColor(entryEditDelim, entryEditDelimColor)
		writeCellEntryTimeSpanDuration(&t, update.After.Start, update.After.End, update.After.Elapsed())
		t.CommitRow()
	}
	if t.Rows() == 0 {
		entryEditNoneColor.Fprintln(stdout, entryEditPrefix, entryEditNoChange)
	} else {
		t.Fprintln(stdout)
	}
}

// PrintEntryList writes a table for a list of entries, grouped by the month,
// week, and day in that order, to STDOUT.
func PrintEntryList(entries []dinkur.Entry) {
	PrintEntryListSearched(entries, "", "")
}

// PrintEntryListSearched writes a table for a list of entries, grouped by the month,
// week, and day in that order, to STDOUT., to STDOUT, as well as highlighting search
// terms (if any).
func PrintEntryListSearched(entries []dinkur.Entry, searchStart, searchEnd string) {
	if len(entries) == 0 {
		tableEmptyColor.Fprintln(stdout, tableEmptyText)
		return
	}
	var reg *regexp.Regexp
	if searchStart != "" || searchEnd != "" {
		var err error
		reg, err = regexp.Compile(fmt.Sprintf("%s(.*?)%s",
			regexp.QuoteMeta(searchStart), regexp.QuoteMeta(searchEnd)))
		if err != nil {
			PrintFatal("Failed to compile highlight regex:", err)
		}
	}
	var t table
	t.SetSpacing("  ")
	t.SetPrefix("  ")
	t.WriteColoredRow(tableHeaderColor, "ID", "NAME", "WEEK", "MONTH", "DAY", "START", "END", "DURATION")
	monthGroups := groupEntries(&entryGroup{groupBy: month{}}, entries)
	for monthGroupIndex, monthGroup := range monthGroups {
		if monthGroupIndex > 0 {
			t.CommitRow() // commit empty delimiting row between different months
		}
		weekGroups := groupEntries(&entryGroup{groupBy: week{}}, monthGroup.getEntries())
		for weekGroupIndex, weekGroup := range weekGroups {
			if weekGroupIndex > 0 {
				t.CommitRow() // commit empty delimiting row between different weeks
			}
			dayGroups := groupEntries(&entryGroup{groupBy: day{}}, weekGroup.getEntries())
			for dayGroupIndex, dayGroup := range dayGroups {
				for entryIndex, entry := range dayGroup.getEntries() {
					writeCellEntryID(&t, entry.ID)
					if reg != nil {
						writeCellEntryNameSearched(&t, entry.Name, reg)
					} else {
						writeCellEntryName(&t, entry.Name)
					}
					firstEntryInDay := entryIndex == 0
					firstDayInWeek := dayGroupIndex == 0
					firstEntryInWeek := firstDayInWeek && firstEntryInDay
					firstWeekInMonth := weekGroupIndex == 0
					firstEntryInMonth := firstWeekInMonth && firstEntryInWeek
					if firstEntryInWeek {
						writeCellWeek(&t, weekGroup)
					} else {
						t.WriteCellColor(tableCellEmptyText, tableCellEmptyColor)
					}
					if firstEntryInMonth {
						writeCellMonth(&t, monthGroup)
					} else {
						t.WriteCellColor(tableCellEmptyText, tableCellEmptyColor)
					}
					if firstEntryInDay {
						writeCellDay(&t, dayGroup)
					} else {
						t.WriteCellColor(tableCellEmptyText, tableCellEmptyColor)
					}
					writeCellEntryStartEnd(&t, entry.Start, entry.End)
					writeCellDuration(&t, entry.Elapsed())

					lastDayInWeekGroup := dayGroupIndex == len(dayGroups)-1
					lastEntryInDayGroup := entryIndex == dayGroup.count()-1
					lastEntryOfWeek := lastDayInWeekGroup && lastEntryInDayGroup
					if lastEntryOfWeek {
						weekSum := sumEntries(weekGroup.getEntries())
						weekDuration := FormatDuration(weekSum.duration)
						cellStr := fmt.Sprintf("Σ Week %s = %s", weekGroup, weekDuration)
						t.WriteCellColor(cellStr, tableWeekSummaryColor)

						lastWeekInMonthGroup := weekGroupIndex == len(weekGroups)-1
						lastEntryOfMonth := lastEntryOfWeek && lastWeekInMonthGroup
						if lastEntryOfMonth {
							t.CommitRow()
							monthSum := sumEntries(monthGroup.getEntries())
							monthDuration := FormatDuration(monthSum.duration)
							cellStr := fmt.Sprintf("Σ Month %s = %s", monthGroup, monthDuration)
							t.WriteColoredRow(
								tableMonthSummaryColor,
								tableCellEmptyText,
								tableCellEmptyText,
								tableCellEmptyText,
								tableCellEmptyText,
								tableCellEmptyText,
								tableCellEmptyText,
								tableCellEmptyText,
								tableCellEmptyText,
								cellStr,
							)
						}
					}
					t.CommitRow()
				}
			}
		}
	}

	sum := sumEntries(entries)
	t.CommitRow() // commit empty delimiting row
	endStr := entryEndNilTextActive
	if sum.end != nil {
		endStr = sum.end.Format(timeFormatShort)
	}
	t.WriteColoredRow(tableSummaryColor,
		tableCellEmptyText, // ID
		fmt.Sprintf("TOTAL: %d entries", len(entries)), // NAME
		tableCellEmptyText,                // WEEK
		tableCellEmptyText,                // MONTH
		tableCellEmptyText,                // DAY
		sum.start.Format(timeFormatShort), // START
		endStr,                            // END
		FormatDuration(sum.duration),      // DURATION
	)
	t.Fprintln(stdout)
}

// PrintEntryListCompact writes a table for a list of entries, grouped by the
// month, week, and day in that order, to STDOUT, as well as highlighting search
// terms (if any). Compacts entries over the same day into one entry, along with
// omitting some other fields: id, name, start, end.
func PrintEntryListCompact(entries []dinkur.Entry) {
	if len(entries) == 0 {
		tableEmptyColor.Fprintln(stdout, tableEmptyText)
		return
	}
	var t table
	t.SetSpacing("  ")
	t.SetPrefix("  ")
	t.WriteColoredRow(tableHeaderColor, "WEEK", "MONTH", "DAY", "DURATION")
	monthGroups := groupEntries(&entryGroup{groupBy: month{}}, entries)
	for monthGroupIndex, monthGroup := range monthGroups {
		if monthGroupIndex > 0 {
			t.CommitRow() // commit empty delimiting row between different months
		}
		weekGroups := groupEntries(&entryGroup{groupBy: week{}}, monthGroup.getEntries())
		for weekGroupIndex, weekGroup := range weekGroups {
			if weekGroupIndex > 0 {
				t.CommitRow() // commit empty delimiting row between different weeks
			}
			dayGroups := groupEntries(&entryGroup{groupBy: day{}}, weekGroup.getEntries())
			for dayGroupIndex, dayGroup := range dayGroups {
				firstDayInWeek := dayGroupIndex == 0
				firstEntryInWeek := firstDayInWeek
				firstWeekInMonth := weekGroupIndex == 0
				firstEntryInMonth := firstWeekInMonth && firstEntryInWeek
				if firstEntryInWeek {
					writeCellWeek(&t, weekGroup)
				} else {
					t.WriteCellColor(tableCellEmptyText, tableCellEmptyColor)
				}
				if firstEntryInMonth {
					writeCellMonth(&t, monthGroup)
				} else {
					t.WriteCellColor(tableCellEmptyText, tableCellEmptyColor)
				}
				writeCellDay(&t, dayGroup)

				daySum := sumEntries(dayGroup.getEntries())
				writeCellDuration(&t, daySum.duration)

				lastDayInWeekGroup := dayGroupIndex == len(dayGroups)-1
				lastEntryOfWeek := lastDayInWeekGroup
				if lastEntryOfWeek {
					weekSum := sumEntries(weekGroup.getEntries())
					weekDuration := FormatDuration(weekSum.duration)
					cellStr := fmt.Sprintf("Σ Week %s = %s", weekGroup, weekDuration)
					t.WriteCellColor(cellStr, tableWeekSummaryColor)

					lastWeekInMonthGroup := weekGroupIndex == len(weekGroups)-1
					lastEntryOfMonth := lastEntryOfWeek && lastWeekInMonthGroup
					if lastEntryOfMonth {
						t.CommitRow()
						monthSum := sumEntries(monthGroup.getEntries())
						monthDuration := FormatDuration(monthSum.duration)
						cellStr := fmt.Sprintf("Σ Month %s = %s", monthGroup, monthDuration)
						t.WriteColoredRow(
							tableMonthSummaryColor,
							tableCellEmptyText, // WEEK
							tableCellEmptyText, // MONTH
							tableCellEmptyText, // DAY
							tableCellEmptyText, // DURATION
							cellStr,
						)
					}
				}
				t.CommitRow()
			}
		}
	}

	sum := sumEntries(entries)
	t.CommitRow() // commit empty delimiting row
	t.WriteColoredRow(tableSummaryColor,
		tableCellEmptyText,           // WEEK
		tableCellEmptyText,           // MONTH
		tableCellEmptyText,           // DAY
		FormatDuration(sum.duration), // DURATION
	)
	t.Fprintln(stdout)
}

// UsageTemplate returns a lightly colored usage template for Cobra.
func UsageTemplate() string {
	var sb strings.Builder
	usageHeaderColor.Fprint(&sb, "Usage:")
	sb.WriteString(`{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

`)
	usageHeaderColor.Fprint(&sb, "Aliases:")
	sb.WriteString(`
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

`)
	usageHeaderColor.Fprint(&sb, "Examples:")
	sb.WriteString(`
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

`)
	usageHeaderColor.Fprint(&sb, "Available Commands:")
	sb.WriteString(`{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

`)
	usageHeaderColor.Fprint(&sb, "Flags:")
	sb.WriteString(`
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

`)
	usageHeaderColor.Fprint(&sb, "Global Flags:")
	sb.WriteString(`
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

`)
	usageHeaderColor.Fprint(&sb, "Additional help topics:")
	sb.WriteString(`{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

`)
	usageHelpColor.Fprint(&sb, `Use "{{.CommandPath}} [command] --help" for more information about a command.`)
	sb.WriteString(`{{end}}`)
	sb.WriteByte('\n')
	return sb.String()
}
