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

package dinkurdb

import (
	"errors"

	"gorm.io/gorm"
)

var (
	ErrAlreadyConnected   = errors.New("client is already connected to database")
	ErrNotConnected       = errors.New("client is not connected to database")
	ErrTaskNameEmpty      = errors.New("task name cannot be empty")
	ErrTaskEndBeforeStart = errors.New("task end date cannot be before start date")
	ErrNotFound           = gorm.ErrRecordNotFound
)

func nilNotFoundError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	return err
}
