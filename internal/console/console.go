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

package console

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/dinkur/dinkur/pkg/dinkur"
	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
)

var (
	stdout          = colorable.NewColorableStdout()
	stderr          = colorable.NewColorableStderr()
	timeFormatLong  = "Jan 02 15:04"
	timeFormatShort = "15:04"
	durationTrunc   = time.Second

	taskIDColor        = color.New(color.FgHiBlack)
	taskLabelColor     = color.New(color.FgWhite, color.Italic)
	taskNameColor      = color.New(color.FgHiYellow, color.Bold)
	taskTimeDelimColor = color.New(color.FgHiBlack)
	taskStartColor     = color.New(color.FgHiGreen)
	taskEndColor       = color.New(color.FgHiGreen)
	taskEndNilColor    = color.New(color.FgGreen, color.Italic)
	taskEndNilText     = "now…"
	taskEndNilTextLen  = utf8.RuneCountInString(taskEndNilText)
	taskDurationColor  = color.New(color.FgCyan)
	taskEditDelimColor = color.New(color.FgHiMagenta)
	taskEditNoneColor  = color.New(color.FgHiBlack, color.Italic)

	debugLabel      = "[DEBUG] "
	debugLabelColor = color.New(color.FgHiBlack, color.Italic)
	debugValueColor = color.New(color.FgHiBlack, color.Italic)

	fatalLabelColor = color.New(color.FgHiRed, color.Bold)
	fatalValueColor = color.New(color.FgRed)

	tableHeaderColor = color.New(color.FgWhite, color.Underline)
)

func PrintTaskWithDuration(label string, task dinkur.Task) {
	var sb strings.Builder
	taskLabelColor.Fprint(&sb, label)
	sb.WriteByte(' ')
	taskIDColor.Fprint(&sb, "#", task.ID)
	sb.WriteByte(' ')
	writeTaskName(&sb, task.Name)
	sb.WriteByte(' ')
	writeTaskTimeSpan(&sb, task.Start, task.End)
	sb.WriteByte(' ')
	writeTaskDuration(&sb, task.Elapsed())
	fmt.Fprintln(stdout, sb.String())
}

func PrintTask(label string, task dinkur.Task) {
	var sb strings.Builder
	taskLabelColor.Fprint(&sb, label)
	sb.WriteByte(' ')
	taskIDColor.Fprint(&sb, "#", task.ID)
	sb.WriteByte(' ')
	writeTaskName(&sb, task.Name)
	sb.WriteByte(' ')
	writeTaskTimeSpan(&sb, task.Start, task.End)
	fmt.Fprintln(stdout, sb.String())
}

func writeTaskName(w io.Writer, name string) {
	taskNameColor.Fprint(w, `"`, name, `"`)
}

func writeTaskTimeSpan(w io.Writer, start time.Time, end *time.Time) {
	today := newDate(time.Now().Date())
	layout := timeFormatShort
	if today != newDate(start.Date()) {
		layout = timeFormatLong
	} else if end != nil && newDate(end.Date()) != today {
		layout = timeFormatLong
	}
	taskStartColor.Fprintf(w, start.Format(layout))
	taskTimeDelimColor.Fprint(w, " - ")
	if end != nil {
		taskEndColor.Fprintf(w, end.Format(layout))
	} else {
		taskEndNilColor.Fprintf(w, taskEndNilText)
	}
}

func newDate(year int, month time.Month, day int) date {
	return date{year, month, day}
}

type date struct {
	year  int
	month time.Month
	day   int
}

func writeTaskDuration(w io.Writer, dur time.Duration) {
	taskTimeDelimColor.Fprint(w, "(")
	taskDurationColor.Fprint(w, dur.Truncate(durationTrunc))
	taskTimeDelimColor.Fprint(w, ")")
}

func PrintDebug(v interface{}) {
	var sb strings.Builder
	debugLabelColor.Fprint(&sb, debugLabel)
	debugValueColor.Fprint(&sb, v)
	fmt.Fprintln(stderr, sb.String())
}

func PrintDebugf(format string, v ...interface{}) {
	var sb strings.Builder
	debugLabelColor.Fprint(&sb, debugLabel)
	debugValueColor.Fprintf(&sb, format, v...)
	fmt.Fprintln(stderr, sb.String())
}

func PrintFatal(label string, v interface{}) {
	var sb strings.Builder
	fatalLabelColor.Fprint(&sb, label)
	sb.WriteByte(' ')
	fatalValueColor.Fprint(&sb, v)
	fmt.Fprintln(stderr, sb.String())
	os.Exit(1)
}

func PrintTaskEdit(update dinkur.UpdatedTask) {
	const editPrefix = "  "
	const editDelim = "   =>   "
	var anyEdit bool
	var sb strings.Builder
	taskLabelColor.Fprint(&sb, "Updated task ")
	taskIDColor.Fprint(&sb, "#", update.Updated.ID)
	sb.WriteByte(' ')
	writeTaskName(&sb, update.Updated.Name)
	taskLabelColor.Fprint(&sb, ":")
	fmt.Fprintln(&sb)
	if update.Old.Name != update.Updated.Name {
		sb.WriteString(editPrefix)
		writeTaskName(&sb, update.Old.Name)
		taskEditDelimColor.Fprint(&sb, editDelim)
		writeTaskName(&sb, update.Updated.Name)
		fmt.Fprintln(&sb)
		anyEdit = true
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
		sb.WriteString(editPrefix)
		writeTaskTimeSpan(&sb, update.Old.Start, update.Old.End)
		sb.WriteByte(' ')
		writeTaskDuration(&sb, update.Old.Elapsed())
		taskEditDelimColor.Fprint(&sb, editDelim)
		writeTaskTimeSpan(&sb, update.Updated.Start, update.Updated.End)
		sb.WriteByte(' ')
		writeTaskDuration(&sb, update.Updated.Elapsed())
		fmt.Fprintln(&sb)
		anyEdit = true
	}
	if !anyEdit {
		sb.WriteString(editPrefix)
		taskEditNoneColor.Fprint(&sb, "No changes were applied.")
		fmt.Fprintln(&sb)
	}
	fmt.Fprint(stdout, sb.String())
}

func PrintTaskList(tasks []dinkur.Task) {
	var t table
	t.SetSpacing("  ")
	t.SetPrefix("  ")
	t.WriteColoredRow(tableHeaderColor, "ID", "NAME", "START", "END", "DUR")
	for _, task := range tasks {
		t.WriteCellWidth(taskIDColor.Sprintf("#%d", task.ID), uintWidth(task.ID)+1)
		t.WriteCellWidth(taskNameColor.Sprintf(`"%s"`, task.Name), utf8.RuneCountInString(task.Name)+2)
		t.WriteCellWidth(taskStartColor.Sprint(task.Start.Format(timeFormatShort)), len(timeFormatShort))
		if task.End != nil {
			t.WriteCellWidth(taskEndColor.Sprint(task.End.Format(timeFormatShort)), len(timeFormatShort))
		} else {
			t.WriteCellWidth(taskEndNilColor.Sprint(taskEndNilText), taskEndNilTextLen)
		}
		dur := task.Elapsed().Truncate(durationTrunc).String()
		t.WriteCellWidth(taskDurationColor.Sprint(dur), len(dur))
		t.CommitRow()
	}
	t.Fprintln(stdout)
}

func uintWidth(i uint) int {
	switch {
	case i < 1e1:
		return 1
	case i < 1e2:
		return 2
	case i < 1e3:
		return 3
	case i < 1e4:
		return 4
	case i < 1e5:
		return 5
	case i < 1e6:
		return 6
	case i < 1e7:
		return 7
	case i < 1e8:
		return 8
	case i < 1e9:
		return 9
	case i < 1e10:
		return 10
	case i < 1e11:
		return 11
	case i < 1e12:
		return 12
	case i < 1e13:
		return 13
	case i < 1e14:
		return 14
	case i < 1e15:
		return 15
	case i < 1e16:
		return 16
	case i < 1e17:
		return 17
	case i < 1e18:
		return 18
	case i < 1e19:
		return 19
	default:
		return 20
	}
}
