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

func writeTimeAutoLayoutColor(w io.Writer, t time.Time, c *color.Color) int {
	today := newDate(time.Now().Date())
	layout := timeFormatShort
	if today != newDate(t.Date()) {
		layout = timeFormatLong
	}
	return writeTimeColor(w, t, layout, c)
}

func writeEntryID(w io.Writer, id uint) int {
	entryIDColor.Fprintf(w, "#%d", id)
	return uintWidth(id) + 1
}

func writeEntryName(w io.Writer, name string) int {
	entryNameColor.Fprintf(w, entryNameFormat, name)
	return 2 + utf8.RuneCountInString(name)
}

func writeEntryNameSearched(w io.Writer, name string, reg *regexp.Regexp) int {
	matches := reg.FindAllStringSubmatchIndex(name, -1)
	const (
		g0Start = 0 // group 0 = full match
		g0End   = 1
		g1Start = 2
		g1End   = 3
	)
	var width int
	width++
	entryNameColor.Fprint(w, entryNameQuote)
	var lastIdx int
	for _, match := range matches {
		nameUntilMatch := name[lastIdx:match[g0Start]]
		width += utf8.RuneCountInString(nameUntilMatch)
		entryNameColor.Fprint(w, nameUntilMatch)

		matchGroup1 := name[match[g1Start]:match[g1End]]
		width += utf8.RuneCountInString(matchGroup1)
		entryNameHighlightColor.Fprint(w, matchGroup1)

		lastIdx = match[g0End]
	}
	nameUntilEnd := name[lastIdx:]
	width += utf8.RuneCountInString(nameUntilEnd)
	entryNameColor.Fprint(w, nameUntilEnd)
	width++
	entryNameColor.Fprint(w, entryNameQuote)
	return width
}

func writeEntryTimeSpanActive(w io.Writer, start time.Time, end *time.Time) int {
	return writeEntryTimeSpan(w, start, end, entryEndNilTextActive)
}

func writeEntryTimeSpanNow(w io.Writer, start time.Time, end *time.Time) int {
	return writeEntryTimeSpan(w, start, end, entryEndNilTextNow)
}

func writeEntryTimeSpan(w io.Writer, start time.Time, end *time.Time, nowStr string) int {
	today := newDate(time.Now().Date())
	layout := timeFormatShort
	if today != newDate(start.Date()) ||
		(end != nil && newDate(end.Date()) != today) {
		// also, if start date != end date, also use long format.
		// This still applies, through transitivity
		layout = timeFormatLong
	}
	startStr := start.Format(layout)
	entryStartColor.Fprintf(w, startStr)
	entryTimeDelimColor.Fprint(w, " - ")
	var (
		endStr   string
		endLen   int
		endColor *color.Color
	)
	if end != nil {
		endStr = end.Format(layout)
		endColor = entryEndColor
		endLen = len(endStr)
	} else {
		endStr = nowStr
		endColor = entryEndNilColor
		endLen = utf8.RuneCountInString(nowStr)
	}
	endColor.Fprintf(w, endStr)
	return len(startStr) + 3 + endLen
}

func writeEntryTimeSpanNowDuration(w io.Writer, start time.Time, end *time.Time, dur time.Duration) int {
	return writeEntryTimeSpanDuration(w, start, end, entryEndNilTextNow, dur)
}

func writeEntryTimeSpanActiveDuration(w io.Writer, start time.Time, end *time.Time, dur time.Duration) int {
	return writeEntryTimeSpanDuration(w, start, end, entryEndNilTextActive, dur)
}

func writeEntryTimeSpanDuration(w io.Writer, start time.Time, end *time.Time, nowStr string, dur time.Duration) int {
	width := writeEntryTimeSpan(w, start, end, nowStr)
	w.Write([]byte{' '})
	width++
	width += writeEntryDurationWithDelim(w, dur)
	return width
}

func writeEntryDurationWithDelim(w io.Writer, dur time.Duration) int {
	entryTimeDelimColor.Fprint(w, "(")
	width := writeEntryDuration(w, dur)
	entryTimeDelimColor.Fprint(w, ")")
	return width + 2
}

func writeEntryDuration(w io.Writer, dur time.Duration) int {
	str := FormatDuration(dur)
	entryDurationColor.Fprint(w, str)
	return len(str)
}
