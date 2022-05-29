// Dinkur the task time tracking utility.
// <https://github.com/dinkur/dinkur>
//
// Copyright (C) 2021 Kalle Fagerberg
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

// Package fuzzytime contains helper functions to parse times in a more fuzzy
// and natural manner, compared to the built-in time.Parse which only attempts
// a single format.
package fuzzytime

import (
	"errors"
	"strings"
	"time"

	"github.com/olebedev/when"
	"github.com/olebedev/when/rules/common"
	"github.com/olebedev/when/rules/en"
)

var (
	// ErrUnknownFormat is returned when no time layout format was matched.
	ErrUnknownFormat = errors.New("unknown time format")
)

var w *when.Parser

func init() {
	w = when.New(nil)
	w.Add(en.All...)
	w.Add(common.All...)
}

// Parse attempts to parse the string literal "now", a delta time, a list of
// known formats, and lastly via the `when` fuzzy parsing package, and returns
// the time on the first match it finds.
func Parse(s string) (time.Time, error) {
	if strings.EqualFold(s, "now") {
		return time.Now(), nil
	}
	if t, ok := ParseDelta(s); ok {
		return t, nil
	}
	if t, err := ParseKnownLayouts(s); err == nil {
		return t, nil
	}
	return ParseWhen(s)
}

var knownLayouts = []string{
	time.RFC3339,
	time.RFC3339Nano,
	time.RFC822,
	time.RFC822Z,
	time.RFC850,
	time.RFC1123,
	time.RFC1123Z,
}

// ParseKnownLayouts attempts to parse the string according to the date formats
// defined in the IETF RFC822, RFC580, RFC1123, or RFC3339.
func ParseKnownLayouts(s string) (time.Time, error) {
	for _, layout := range knownLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, ErrUnknownFormat
}

// ParseWhen performs a fuzzy time parsing via the `when` package.
func ParseWhen(s string) (time.Time, error) {
	r, err := w.Parse(s, time.Now().Truncate(time.Second))
	if err != nil {
		return time.Time{}, err
	}
	if r == nil {
		return time.Time{}, ErrUnknownFormat
	}
	return r.Time, nil
}

// ParseDelta attempts to parse the string as a time.Duration if it is prefixed
// with a sign ("+" or "-"), and adds that to the current time.
func ParseDelta(s string) (time.Time, bool) {
	if len(s) < 3 {
		return time.Time{}, false
	}
	if s[0] != '+' && s[0] != '-' {
		return time.Time{}, false
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return time.Time{}, false
	}
	return time.Now().Add(d), true
}
