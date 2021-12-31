// Dinkur the task time tracking utility.
// <https://github.com/dinkur/dinkur>
//
// Copyright (C) 2021 Kalle Fagerberg
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

// Package license contains license snippets that can be used to print the
// full GNU GPL license to the console, for example.
package license

import (
	_ "embed"
)

// Header is the copyright and license header text meant to be used in
// in the CLI helper text.
var Header = `dinkur Copyright (C) 2021 Kalle Fagerberg

  License GPLv3+: GNU GPL version 3 or later <https://gnu.org/licenses/gpl.html>
  This program comes with ABSOLUTELY NO WARRANTY; for details run 'dinkur --license-w'.
  This is free software, and you are welcome to redistribute it
  under certain conditions; run 'dinkur --license-c' for details.
`

// Conditions is the full terms and conditions section of the GNU GPL 3.0
// license text.
//
//go:embed license_conditions.txt
var Conditions string

// Warranty is the warranty section of the GNU GPL 3.0 license text.
//
//go:embed license_warranty.txt
var Warranty string
