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

	"github.com/dinkur/dinkur/pkg/dinkurdb"
	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
)

var (
	stdout        = colorable.NewColorableStdout()
	stderr        = colorable.NewColorableStderr()
	timeFormat    = "15:04"
	durationTrunc = time.Second

	taskIDColor        = color.New(color.FgHiBlack)
	taskLabelColor     = color.New(color.FgWhite, color.Italic)
	taskNameColor      = color.New(color.FgHiYellow, color.Bold)
	taskTimeDelimColor = color.New(color.FgHiBlack)
	taskStartColor     = color.New(color.FgHiGreen)
	taskEndColor       = color.New(color.FgHiGreen)
	taskEndNilColor    = color.New(color.FgGreen, color.Italic)
	taskDurationColor  = color.New(color.FgCyan)
	taskEditDelimColor = color.New(color.FgHiMagenta)
	taskEditNoneColor  = color.New(color.FgHiBlack, color.Italic)

	fatalLabelColor = color.New(color.FgHiRed, color.Bold)
	fatalValueColor = color.New(color.FgRed)
)

func PrintTaskWithDuration(label string, task dinkurdb.Task) {
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

func PrintTask(label string, task dinkurdb.Task) {
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
	taskStartColor.Fprintf(w, start.Format(timeFormat))
	taskTimeDelimColor.Fprint(w, " - ")
	if end != nil {
		taskEndColor.Fprintf(w, end.Format(timeFormat))
	} else {
		taskEndNilColor.Fprintf(w, "now…")
	}
}

func writeTaskDuration(w io.Writer, dur time.Duration) {
	taskTimeDelimColor.Fprint(w, "(")
	taskDurationColor.Fprint(w, dur.Truncate(durationTrunc))
	taskTimeDelimColor.Fprint(w, ")")
}

func PrintFatal(label string, v interface{}) {
	var sb strings.Builder
	fatalLabelColor.Fprint(&sb, label)
	sb.WriteByte(' ')
	fatalValueColor.Fprint(&sb, v)
	fmt.Fprintln(stderr, sb.String())
	os.Exit(1)
}

func PrintTaskEdit(update dinkurdb.UpdatedTask) {
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
