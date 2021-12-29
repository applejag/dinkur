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
	"strings"
	"time"

	"github.com/dinkur/dinkur/pkg/dinkurdb"
	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
)

var (
	stdout        = colorable.NewColorableStdout()
	timeFormat    = "15:04"
	durationTrunc = time.Second

	taskLabelColor     = color.New(color.FgWhite, color.Italic)
	taskNameColor      = color.New(color.FgHiYellow, color.Bold)
	taskTimeDelimColor = color.New(color.FgHiBlack)
	taskStartColor     = color.New(color.FgHiGreen)
	taskEndColor       = color.New(color.FgHiGreen)
	taskEndNilColor    = color.New(color.FgGreen, color.Italic)
	taskDurationColor  = color.New(color.FgCyan)
)

func PrintTaskWithDuration(label string, task dinkurdb.Task) {
	sb := prepareTaskString(label, task)
	taskTimeDelimColor.Fprint(sb, ", ")
	taskDurationColor.Fprint(sb, task.Elapsed().Truncate(durationTrunc))
	taskTimeDelimColor.Fprint(sb, " elapsed)")
	fmt.Fprintln(stdout, sb.String())
}

func PrintTask(label string, task dinkurdb.Task) {
	sb := prepareTaskString(label, task)
	taskTimeDelimColor.Fprint(sb, ")")
	fmt.Fprintln(stdout, sb.String())
}

func prepareTaskString(label string, task dinkurdb.Task) *strings.Builder {
	var sb strings.Builder
	taskLabelColor.Fprint(&sb, label)
	sb.WriteByte(' ')
	taskNameColor.Fprint(&sb, `"`, task.Name, `"`)
	sb.WriteByte(' ')
	taskTimeDelimColor.Fprint(&sb, "(")
	taskStartColor.Fprintf(&sb, task.Start.Format(timeFormat))
	taskTimeDelimColor.Fprint(&sb, " => ")
	if task.End != nil {
		taskEndColor.Fprintf(&sb, task.End.Format(timeFormat))
	} else {
		taskEndNilColor.Fprintf(&sb, "now…")
	}
	return &sb
}
