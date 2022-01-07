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

	taskIDColor              = color.New(color.FgHiBlack)
	taskLabelColor           = color.New(color.FgWhite, color.Italic)
	taskNameColor            = color.New(color.FgYellow)
	taskNameHighlightColor   = color.New(color.FgHiYellow, color.Underline)
	taskNameHighlightReplace = taskNameHighlightColor.Sprint("$1")
	taskNameQuote            = "`"
	taskNameFormat           = taskNameQuote + "%s" + taskNameQuote
	taskTimeDelimColor       = color.New(color.FgHiBlack)
	taskDateColor            = color.New(color.FgGreen)
	taskStartColor           = color.New(color.FgGreen)
	taskEndColor             = color.New(color.FgGreen)
	taskEndNilColor          = color.New(color.FgHiBlack, color.Italic)
	taskEndNilTextNow        = "now…"
	taskEndNilTextActive     = "active…"
	taskDurationColor        = color.New(color.FgCyan)
	taskEditDelimColor       = color.New(color.FgHiMagenta)
	taskEditNoneColor        = color.New(color.FgHiBlack, color.Italic)

	taskEditPrefix  = "  "
	taskEditSpacing = "   "
	taskEditDelim   = "=>"

	fatalLabelColor = color.New(color.FgHiRed, color.Bold)
	fatalValueColor = color.New(color.FgRed)

	tableEmptyColor     = color.New(color.FgHiBlack, color.Italic)
	tableEmptyText      = "No results to display."
	tableHeaderColor    = color.New(color.FgHiBlack)
	tableSummaryColor   = color.New(color.FgHiBlack, color.Italic)
	tableCellEmptyText  = "-"
	tableCellEmptyColor = color.New(color.FgHiBlack)

	usageHeaderColor = color.New(color.FgYellow, color.Underline, color.Italic)
	usageHelpColor   = color.New(color.FgHiBlack, color.Italic)

	promptWarnIconColor = color.New(color.FgHiRed, color.Bold)
	promptWarnIconText  = "!"
	promptErrorColor    = color.New(color.FgRed)
)

// LabelledTask holds a string label and a task. Used when printing multiple
// labelled tasks together.
type LabelledTask struct {
	Label      string
	Task       dinkur.Task
	NoDuration bool
}

// PrintTaskLabel writes a label string followed by a formatted task to STDOUT.
func PrintTaskLabel(labelled LabelledTask) {
	var t table
	t.SetSpacing("  ")
	t.WriteColoredRow(tableHeaderColor, "", "ID", "NAME", "START", "END", "DURATION")
	writeCellsLabelledTask(&t, labelled)
	t.CommitRow()
	t.Fprintln(stdout)
}

// PrintTaskLabelSlice writes a table of label strings followed by a formatted
// task to STDOUT.
func PrintTaskLabelSlice(slice []LabelledTask) {
	var t table
	t.SetSpacing("  ")
	t.WriteColoredRow(tableHeaderColor, "", "ID", "NAME", "START", "END", "DURATION")
	for _, lbl := range slice {
		writeCellsLabelledTask(&t, lbl)
		t.CommitRow()
	}
	t.Fprintln(stdout)
}

// PrintFatal writes a label and some error value to STDERR and then exits the
// application with status code 1.
func PrintFatal(label string, v interface{}) {
	var sb strings.Builder
	fatalLabelColor.Fprint(&sb, label)
	sb.WriteByte(' ')
	fatalValueColor.Fprint(&sb, v)
	fmt.Fprintln(stderr, sb.String())
	os.Exit(1)
}

// PrintTaskEdit writes a formatted task and highlights any edits made to it,
// by diffing the before and after tasks, to STDOUT.
func PrintTaskEdit(update dinkur.UpdatedTask) {
	var sb strings.Builder
	taskLabelColor.Fprint(&sb, "Updated task ")
	taskIDColor.Fprint(&sb, "#", update.Updated.ID)
	sb.WriteByte(' ')
	writeTaskName(&sb, update.Updated.Name)
	taskLabelColor.Fprint(&sb, ":")
	fmt.Fprintln(stdout, sb.String())

	var t table
	t.SetPrefix(taskEditPrefix)
	t.SetSpacing(taskEditSpacing)
	if update.Old.Name != update.Updated.Name {
		writeCellTaskName(&t, update.Old.Name)
		t.WriteCellColor(taskEditDelim, taskEditDelimColor)
		writeCellTaskName(&t, update.Updated.Name)
		t.CommitRow()
	}
	var (
		oldStartUnix = update.Old.Start.UnixMilli()
		oldEndUnix   int64
		newStartUnix = update.Updated.Start.UnixMilli()
		newEndUnix   int64
	)
	if update.Old.End != nil {
		oldEndUnix = update.Old.End.Unix()
	}
	if update.Updated.End != nil {
		newEndUnix = update.Updated.End.Unix()
	}
	if oldStartUnix != newStartUnix || oldEndUnix != newEndUnix {
		writeCellTaskTimeSpanDuration(&t, update.Old.Start, update.Old.End, update.Old.Elapsed())
		t.WriteCellColor(taskEditDelim, taskEditDelimColor)
		writeCellTaskTimeSpanDuration(&t, update.Updated.Start, update.Updated.End, update.Updated.Elapsed())
		t.CommitRow()
	}
	if t.Rows() == 0 {
		taskEditNoneColor.Fprint(stdout, taskEditPrefix, "No changes were applied.")
		fmt.Fprintln(&sb)
	} else {
		t.Fprintln(stdout)
	}
}

// PrintTaskList writes a table for a list of tasks, grouped by the date
// (year, month, day), to STDOUT.
func PrintTaskList(tasks []dinkur.Task) {
	PrintTaskListSearched(tasks, "", "")
}

// PrintTaskListSearched writes a table for a list of tasks, grouped by the date
// (year, month, day), to STDOUT, as well as highlighting search terms (if any).
func PrintTaskListSearched(tasks []dinkur.Task, searchStart, searchEnd string) {
	if len(tasks) == 0 {
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
	t.WriteColoredRow(tableHeaderColor, "ID", "NAME", "DAY", "START", "END", "DURATION")
	for i, group := range groupTasksByDate(tasks) {
		if i > 0 {
			t.CommitRow() // commit empty delimiting row
		}
		for i, task := range group.tasks {
			writeCellTaskID(&t, task.ID)
			if reg != nil {
				writeCellTaskNameSearched(&t, task.Name, reg)
			} else {
				writeCellTaskName(&t, task.Name)
			}
			if i == 0 {
				writeCellDate(&t, group.date)
			} else {
				t.WriteCellColor(tableCellEmptyText, tableCellEmptyColor)
			}
			writeCellTaskStartEnd(&t, task.Start, task.End)
			writeCellDuration(&t, task.Elapsed())
			t.CommitRow()
		}
	}
	sum := sumTasks(tasks)
	t.CommitRow() // commit empty delimiting row
	endStr := taskEndNilTextActive
	if sum.end != nil {
		endStr = sum.end.Format(timeFormatShort)
	}
	t.WriteColoredRow(tableSummaryColor,
		tableCellEmptyText,                         // ID
		fmt.Sprintf("TOTAL: %d tasks", len(tasks)), // NAME
		tableCellEmptyText,                         // DAY
		sum.start.Format(timeFormatShort),          // START
		endStr,                                     // END
		FormatDuration(sum.duration),               // DURATION
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
