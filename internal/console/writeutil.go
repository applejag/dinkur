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

package console

import (
	"io"
	"regexp"
	"time"
	"unicode/utf8"

	"github.com/fatih/color"
)

func writeTimeColor(w io.Writer, t time.Time, layout string, c *color.Color) int {
	formatted := t.Format(layout)
	c.Fprintf(w, formatted)
	return len(formatted)
}

func writeTaskID(w io.Writer, id uint) int {
	taskIDColor.Fprintf(w, "#%d", id)
	return uintWidth(id) + 1
}

func writeTaskName(w io.Writer, name string) int {
	taskNameColor.Fprintf(w, taskNameFormat, name)
	return 2 + utf8.RuneCountInString(name)
}

func writeTaskNameSearched(w io.Writer, name string, reg *regexp.Regexp) int {
	matches := reg.FindAllStringSubmatchIndex(name, -1)
	const (
		g0Start = 0 // group 0 = full match
		g0End   = 1
		g1Start = 2
		g1End   = 3
	)
	var width int
	width++
	taskNameColor.Fprint(w, taskNameQuote)
	var lastIdx int
	for _, match := range matches {
		nameUntilMatch := name[lastIdx:match[g0Start]]
		width += utf8.RuneCountInString(nameUntilMatch)
		taskNameColor.Fprint(w, nameUntilMatch)

		matchGroup1 := name[match[g1Start]:match[g1End]]
		width += utf8.RuneCountInString(matchGroup1)
		taskNameHighlightColor.Fprint(w, matchGroup1)

		lastIdx = match[g0End]
	}
	nameUntilEnd := name[lastIdx:]
	width += utf8.RuneCountInString(nameUntilEnd)
	taskNameColor.Fprint(w, nameUntilEnd)
	width++
	taskNameColor.Fprint(w, taskNameQuote)
	return width
}

func writeTaskTimeSpanActive(w io.Writer, start time.Time, end *time.Time) int {
	return writeTaskTimeSpan(w, start, end, taskEndNilTextActive)
}

func writeTaskTimeSpanNow(w io.Writer, start time.Time, end *time.Time) int {
	return writeTaskTimeSpan(w, start, end, taskEndNilTextNow)
}

func writeTaskTimeSpan(w io.Writer, start time.Time, end *time.Time, nowStr string) int {
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
	var (
		endStr   string
		endLen   int
		endColor *color.Color
	)
	if end != nil {
		endStr = end.Format(layout)
		endColor = taskEndColor
		endLen = len(endStr)
	} else {
		endStr = nowStr
		endColor = taskEndNilColor
		endLen = utf8.RuneCountInString(nowStr)
	}
	endColor.Fprintf(w, endStr)
	return len(startStr) + 3 + endLen
}

func writeTaskTimeSpanNowDuration(w io.Writer, start time.Time, end *time.Time, dur time.Duration) int {
	return writeTaskTimeSpanDuration(w, start, end, taskEndNilTextNow, dur)
}

func writeTaskTimeSpanActiveDuration(w io.Writer, start time.Time, end *time.Time, dur time.Duration) int {
	return writeTaskTimeSpanDuration(w, start, end, taskEndNilTextActive, dur)
}

func writeTaskTimeSpanDuration(w io.Writer, start time.Time, end *time.Time, nowStr string, dur time.Duration) int {
	width := writeTaskTimeSpan(w, start, end, nowStr)
	w.Write([]byte{' '})
	width++
	width += writeTaskDurationWithDelim(w, dur)
	return width
}

func writeTaskDurationWithDelim(w io.Writer, dur time.Duration) int {
	taskTimeDelimColor.Fprint(w, "(")
	width := writeTaskDuration(w, dur)
	taskTimeDelimColor.Fprint(w, ")")
	return width + 2
}

func writeTaskDuration(w io.Writer, dur time.Duration) int {
	str := FormatDuration(dur)
	taskDurationColor.Fprint(w, str)
	return len(str)
}
