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

package fromgrpc

import (
	"errors"
	"fmt"

	dinkurapiv1 "github.com/dinkur/dinkur/api/dinkurapi/v1"
	"github.com/dinkur/dinkur/pkg/conv"
	"github.com/dinkur/dinkur/pkg/dinkur"
)

// Errors that are specific to converting gRPC entries to Go.
var (
	ErrUnexpectedNilEntry  = errors.New("unexpected nil entry")
	ErrUnexpectedNilStatus = errors.New("unexpected nil status")
)

// EntryPtr converts a gRPC entry to a Go entry.
func EntryPtr(entry *dinkurapiv1.Entry) (*dinkur.Entry, error) {
	if entry == nil {
		return nil, nil
	}
	id, err := conv.Uint64ToUint(entry.Id)
	if err != nil {
		return nil, fmt.Errorf("convert entry ID: %w", err)
	}
	return &dinkur.Entry{
		CommonFields: dinkur.CommonFields{
			TimeFields: dinkur.TimeFields{
				CreatedAt: TimeOrZero(entry.Created),
				UpdatedAt: TimeOrZero(entry.Updated),
			},
			ID: id,
		},
		Name:  entry.Name,
		Start: TimeOrZero(entry.Start),
		End:   TimePtr(entry.End),
	}, nil
}

// EntryPtrNoNil converts a gRPC entry to a Go entry, or error if nil.
func EntryPtrNoNil(entry *dinkurapiv1.Entry) (dinkur.Entry, error) {
	t, err := EntryPtr(entry)
	if err != nil {
		return dinkur.Entry{}, err
	}
	if t == nil {
		return dinkur.Entry{}, ErrUnexpectedNilEntry
	}
	return *t, nil
}

// EntrySlice converts a slice of gRPC entries to Go entries. Nils are skipped.
func EntrySlice(slice []*dinkurapiv1.Entry) ([]dinkur.Entry, error) {
	entries := make([]dinkur.Entry, 0, len(slice))
	for _, t := range slice {
		t2, err := EntryPtr(t)
		if t2 == nil {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("entry #%d %q: %w", t.Id, t.Name, err)
		}
		entries = append(entries, *t2)
	}
	return entries, nil
}
