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
	"io"
	"time"
	"unicode/utf8"

	"github.com/fatih/color"
)

func writeTimeColor(w io.Writer, t time.Time, layout string, c *color.Color) int {
	formatted := t.Format(layout)
	c.Fprintf(w, formatted)
	return len(formatted)
}

func writeTaskName(w io.Writer, name string) int {
	taskNameColor.Fprintf(w, `"%s"`, name)
	return 2 + utf8.RuneCountInString(name)
}

func writeTaskTimeSpan(w io.Writer, start time.Time, end *time.Time) int {
	today := newDate(time.Now().Date())
	layout := timeFormatShort
	if today != newDate(start.Date()) ||
		(end != nil && newDate(end.Date()) != today) {
		// also, if start date != end date, also use long format.
		// This still applies, through transitivity
		layout = timeFormatLong
	}
	startStr := start.Format(layout)
	taskStartColor.Fprintf(w, startStr)
	taskTimeDelimColor.Fprint(w, " - ")
	endStr := taskEndNilText
	endColor := taskEndNilColor
	endLen := taskEndNilTextLen
	if end != nil {
		endStr = end.Format(layout)
		endColor = taskEndColor
		endLen = len(endStr)
	}
	endColor.Fprintf(w, endStr)
	return len(startStr) + 3 + endLen
}

func writeTaskDuration(w io.Writer, dur time.Duration) int {
	taskTimeDelimColor.Fprint(w, "(")
	str := FormatDuration(dur)
	taskDurationColor.Fprint(w, str)
	taskTimeDelimColor.Fprint(w, ")")
	return len(str) + 2
}
