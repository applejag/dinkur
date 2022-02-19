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

package dinkurdb

import (
	"context"
	"errors"

	"github.com/dinkur/dinkur/pkg/dinkur"
)

// StreamStatus is a dummy implementation of the dinkur.Client that only
// returns the "client is nil" error.
func (*client) StreamStatus(context.Context) (<-chan dinkur.StreamedStatus, error) {
	return nil, errors.New("not implemented")
}

// SetStatus is a dummy implementation of the dinkur.Client that only
// returns the "client is nil" error.
func (*client) SetStatus(context.Context, dinkur.EditStatus) error {
	return errors.New("not implemented")
}

// GetStatus is a dummy implementation of the dinkur.Client that only
// returns the "client is nil" error.
func (*client) GetStatus(context.Context) (dinkur.Status, error) {
	return dinkur.Status{}, errors.New("not implemented")
}
