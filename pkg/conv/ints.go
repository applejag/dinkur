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

package conv

import (
	"fmt"
	"math"
)

// Errors that are specific to model conversion.
var (
	ErrUintTooLarge = fmt.Errorf("unsigned int value is too large, maximum: %d", uint64(math.MaxUint))
)

// Uint64ToUint converts a uint64 to uint, and errors if uint64 value is larger
// than what the uint type supports, such as having a value bigger than 2^32-1
// in a 32-bit build of Dinkur.
func Uint64ToUint(v uint64) (uint, error) {
	if v > math.MaxUint {
		return 0, ErrUintTooLarge
	}
	return uint(v), nil
}
