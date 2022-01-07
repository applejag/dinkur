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
	"regexp"
	"strings"
	"time"

	"github.com/fatih/color"
)

func writeCellsLabelledTask(t *table, labelled LabelledTask) {
	t.WriteCellColor(labelled.Label, taskLabelColor)
	writeCellTaskID(t, labelled.Task.ID)
	writeCellTaskName(t, labelled.Task.Name)
	writeCellTaskStartEnd(t, labelled.Task.Start, labelled.Task.End)
	if labelled.NoDuration {
		t.WriteCellColor(tableCellEmptyText, tableCellEmptyColor)
	} else {
		writeCellDuration(t, labelled.Task.Elapsed())
	}
}

func writeCellTaskID(t *table, id uint) {
	var sb strings.Builder
	width := writeTaskID(&sb, id)
	t.WriteCellWidth(sb.String(), width)
}

func writeCellTaskName(t *table, name string) {
	var sb strings.Builder
	width := writeTaskName(&sb, name)
	t.WriteCellWidth(sb.String(), width)
}

func writeCellTaskNameSearched(t *table, name string, reg *regexp.Regexp) {
	var sb strings.Builder
	width := writeTaskNameSearched(&sb, name, reg)
	t.WriteCellWidth(sb.String(), width)
}

func writeCellDate(t *table, d date) {
	dateStr := d.String()
	t.WriteCellColor(dateStr, taskDateColor)
}

func writeCellTimeColor(t *table, ti time.Time, layout string, c *color.Color) {
	var sb strings.Builder
	width := writeTimeColor(&sb, ti, layout, c)
	t.WriteCellWidth(sb.String(), width)
}

func writeCellTaskTimeSpan(t *table, start time.Time, end *time.Time) {
	var sb strings.Builder
	width := writeTaskTimeSpanActive(&sb, start, end)
	t.WriteCellWidth(sb.String(), width)
}

func writeCellTaskTimeSpanDuration(t *table, start time.Time, end *time.Time, dur time.Duration) {
	var sb strings.Builder
	width := writeTaskTimeSpanActiveDuration(&sb, start, end, dur)
	t.WriteCellWidth(sb.String(), width)
}

func writeCellTaskStartEnd(t *table, start time.Time, end *time.Time) {
	writeCellTimeColor(t, start, timeFormatShort, taskStartColor)
	if end != nil {
		var endLayout = timeFormatShort
		if newDate(end.Date()) != newDate(start.Date()) {
			endLayout = timeFormatLong
		}
		writeCellTimeColor(t, *end, endLayout, taskEndColor)
	} else {
		t.WriteCellColor(taskEndNilTextActive, taskEndNilColor)
	}
}

func writeCellDuration(t *table, d time.Duration) {
	var sb strings.Builder
	width := writeTaskDuration(&sb, d)
	t.WriteCellWidth(sb.String(), width)
}
