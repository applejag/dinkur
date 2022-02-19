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

package fromdb

import (
	"github.com/dinkur/dinkur/pkg/conv"
	"github.com/dinkur/dinkur/pkg/dbmodel"
	"github.com/dinkur/dinkur/pkg/dinkur"
)

// Status converts a DB status model to a dinkur status model.
func Status(status dbmodel.Status) dinkur.Status {
	return dinkur.Status{
		TimeFields: TimeFields(status.CommonFields),
		AFKSince:   conv.TimePtrLocal(status.AFKSince),
		BackSince:  conv.TimePtrLocal(status.BackSince),
	}
}
