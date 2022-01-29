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

	"github.com/dinkur/dinkur/pkg/dinkur"
)

func (*client) StreamAlert(context.Context) (<-chan dinkur.StreamedAlert, error) {
	return nil, ErrAlerterNotSupported
}

func (*client) GetAlertList(context.Context) ([]dinkur.Alert, error) {
	return nil, ErrAlerterNotSupported
}

func (*client) DeleteAlert(context.Context, uint) (dinkur.Alert, error) {
	return nil, ErrAlerterNotSupported
}
