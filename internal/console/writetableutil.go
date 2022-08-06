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
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/fatih/color"
)

func writeCellsLabelledEntry(t *table, labelled LabelledEntry) {
	t.WriteCellColor(labelled.Label, entryLabelColor)
	writeCellEntryID(t, labelled.Entry.ID)
	writeCellEntryName(t, labelled.Entry.Name)
	writeCellEntryStartEnd(t, labelled.Entry.Start, labelled.Entry.End)
	if labelled.NoDuration {
		t.WriteCellColor(tableCellEmptyText, tableCellEmptyColor)
	} else {
		writeCellDuration(t, labelled.Entry.Elapsed())
	}
}

func writeCellEntryID(t *table, id uint) {
	var sb strings.Builder
	width := writeEntryID(&sb, id)
	t.WriteCellWidth(sb.String(), width)
}

func writeCellEntryName(t *table, name string) {
	var sb strings.Builder
	width := writeEntryName(&sb, name)
	t.WriteCellWidth(sb.String(), width)
}

func writeCellEntryNameSearched(t *table, name string, reg *regexp.Regexp) {
	var sb strings.Builder
	width := writeEntryNameSearched(&sb, name, reg)
	t.WriteCellWidth(sb.String(), width)
}

func writeCellMonth(t *table, m fmt.Stringer) {
	monthStr := m.String()
	t.WriteCellColor(monthStr, entryMonthColor)
}

func writeCellWeek(t *table, w fmt.Stringer) {
	weekStr := w.String()
	t.WriteCellColor(weekStr, entryWeekColor)
}

func writeCellDay(t *table, d fmt.Stringer) {
	dayStr := d.String()
	t.WriteCellColor(dayStr, entryDayColor)
}

func writeCellTimeColor(t *table, ti time.Time, layout string, c *color.Color) {
	var sb strings.Builder
	width := writeTimeColor(&sb, ti, layout, c)
	t.WriteCellWidth(sb.String(), width)
}

func writeCellEntryTimeSpan(t *table, start time.Time, end *time.Time) {
	var sb strings.Builder
	width := writeEntryTimeSpanActive(&sb, start, end)
	t.WriteCellWidth(sb.String(), width)
}

func writeCellEntryTimeSpanDuration(t *table, start time.Time, end *time.Time, dur time.Duration) {
	var sb strings.Builder
	width := writeEntryTimeSpanActiveDuration(&sb, start, end, dur)
	t.WriteCellWidth(sb.String(), width)
}

func writeCellEntryStartEnd(t *table, start time.Time, end *time.Time) {
	writeCellTimeColor(t, start, timeFormatShort, entryStartColor)
	if end != nil {
		var endLayout = timeFormatShort
		d := day{}
		if d.new(*end) != d.new(start) {
			endLayout = timeFormatLong
		}
		writeCellTimeColor(t, *end, endLayout, entryEndColor)
	} else {
		t.WriteCellColor(entryEndNilTextActive, entryEndNilColor)
	}
}

func writeCellDuration(t *table, d time.Duration) {
	var sb strings.Builder
	width := writeEntryDuration(&sb, d)
	t.WriteCellWidth(sb.String(), width)
}
