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

package dinkurclient

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/dinkur/dinkur/pkg/dinkur"

	dinkurapiv1 "github.com/dinkur/dinkur/api/dinkurapi/v1"
)

func (c *client) GetEntry(ctx context.Context, id uint) (dinkur.Entry, error) {
	if err := c.assertConnected(); err != nil {
		return dinkur.Entry{}, err
	}
	res, err := c.entryer.GetEntry(ctx, &dinkurapiv1.GetEntryRequest{
		Id: uint64(id),
	})
	if err != nil {
		return dinkur.Entry{}, convError(err)
	}
	if res == nil {
		return dinkur.Entry{}, ErrResponseIsNil
	}
	entry, err := convEntryPtrNoNil(res.Entry)
	if err != nil {
		return dinkur.Entry{}, convError(err)
	}
	return entry, nil
}

func (c *client) GetEntryList(ctx context.Context, search dinkur.SearchEntry) ([]dinkur.Entry, error) {
	if err := c.assertConnected(); err != nil {
		return nil, err
	}
	req := dinkurapiv1.GetEntryListRequest{
		Start:              convTimePtr(search.Start),
		End:                convTimePtr(search.End),
		Limit:              uint64(search.Limit),
		Shorthand:          convShorthand(search.Shorthand),
		NameFuzzy:          search.NameFuzzy,
		NameHighlightStart: search.NameHighlightStart,
		NameHighlightEnd:   search.NameHighlightEnd,
	}
	res, err := c.entryer.GetEntryList(ctx, &req)
	if err != nil {
		return nil, convError(err)
	}
	if res == nil {
		return nil, ErrResponseIsNil
	}
	entries, err := convEntrySlice(res.Entries)
	if err != nil {
		return nil, convError(err)
	}
	return entries, nil
}

func (c *client) UpdateEntry(ctx context.Context, edit dinkur.EditEntry) (dinkur.UpdatedEntry, error) {
	if err := c.assertConnected(); err != nil {
		return dinkur.UpdatedEntry{}, err
	}
	res, err := c.entryer.UpdateEntry(ctx, &dinkurapiv1.UpdateEntryRequest{
		IdOrZero:           uint64(edit.IDOrZero),
		Name:               convStringPtr(edit.Name),
		Start:              convTimePtr(edit.Start),
		End:                convTimePtr(edit.End),
		AppendName:         edit.AppendName,
		StartAfterIdOrZero: uint64(edit.StartAfterIDOrZero),
		EndBeforeIdOrZero:  uint64(edit.EndBeforeIDOrZero),
		StartAfterLast:     edit.StartAfterLast,
	})
	if err != nil {
		return dinkur.UpdatedEntry{}, convError(err)
	}
	if res == nil {
		return dinkur.UpdatedEntry{}, ErrResponseIsNil
	}
	entryBefore, err := convEntryPtrNoNil(res.Before)
	if err != nil {
		return dinkur.UpdatedEntry{}, fmt.Errorf("entry before: %w", convError(err))
	}
	entryAfter, err := convEntryPtrNoNil(res.After)
	if err != nil {
		return dinkur.UpdatedEntry{}, fmt.Errorf("entry after: %w", convError(err))
	}
	return dinkur.UpdatedEntry{
		Before: entryBefore,
		After:  entryAfter,
	}, nil
}

func (c *client) DeleteEntry(ctx context.Context, id uint) (dinkur.Entry, error) {
	if err := c.assertConnected(); err != nil {
		return dinkur.Entry{}, err
	}
	res, err := c.entryer.DeleteEntry(ctx, &dinkurapiv1.DeleteEntryRequest{
		Id: uint64(id),
	})
	if err != nil {
		return dinkur.Entry{}, convError(err)
	}
	if res == nil {
		return dinkur.Entry{}, ErrResponseIsNil
	}
	entry, err := convEntryPtrNoNil(res.DeletedEntry)
	if err != nil {
		return dinkur.Entry{}, convError(err)
	}
	return entry, nil
}

func (c *client) CreateEntry(ctx context.Context, entry dinkur.NewEntry) (dinkur.StartedEntry, error) {
	if err := c.assertConnected(); err != nil {
		return dinkur.StartedEntry{}, err
	}
	res, err := c.entryer.CreateEntry(ctx, &dinkurapiv1.CreateEntryRequest{
		Name:               entry.Name,
		Start:              convTimePtr(entry.Start),
		End:                convTimePtr(entry.End),
		StartAfterIdOrZero: uint64(entry.StartAfterIDOrZero),
		EndBeforeIdOrZero:  uint64(entry.EndBeforeIDOrZero),
		StartAfterLast:     entry.StartAfterLast,
	})
	if err != nil {
		return dinkur.StartedEntry{}, convError(err)
	}
	if res == nil {
		return dinkur.StartedEntry{}, ErrResponseIsNil
	}
	prevEntry, err := convEntryPtr(res.PreviouslyActiveEntry)
	if err != nil {
		return dinkur.StartedEntry{}, fmt.Errorf("stopped entry: %w", convError(err))
	}
	newEntry, err := convEntryPtrNoNil(res.CreatedEntry)
	if err != nil {
		return dinkur.StartedEntry{}, fmt.Errorf("created entry: %w", convError(err))
	}
	return dinkur.StartedEntry{
		Stopped: prevEntry,
		Started: newEntry,
	}, nil
}

func (c *client) GetActiveEntry(ctx context.Context) (*dinkur.Entry, error) {
	if err := c.assertConnected(); err != nil {
		return nil, err
	}
	res, err := c.entryer.GetActiveEntry(ctx, &dinkurapiv1.GetActiveEntryRequest{})
	if err != nil {
		return nil, convError(err)
	}
	if res == nil {
		return nil, ErrResponseIsNil
	}
	entry, err := convEntryPtr(res.ActiveEntry)
	if err != nil {
		return nil, convError(err)
	}
	return entry, nil
}

func (c *client) StopActiveEntry(ctx context.Context, endTime time.Time) (*dinkur.Entry, error) {
	if err := c.assertConnected(); err != nil {
		return nil, err
	}
	res, err := c.entryer.StopActiveEntry(ctx, &dinkurapiv1.StopActiveEntryRequest{
		End: convTimePtr(&endTime),
	})
	if err != nil {
		return nil, convError(err)
	}
	if res == nil {
		return nil, ErrResponseIsNil
	}
	entry, err := convEntryPtr(res.StoppedEntry)
	if err != nil {
		return nil, convError(err)
	}
	return entry, nil
}

func (c *client) StreamEntry(ctx context.Context) (<-chan dinkur.StreamedEntry, error) {
	if err := c.assertConnected(); err != nil {
		return nil, err
	}
	stream, err := c.entryer.StreamEntry(ctx, &dinkurapiv1.StreamEntryRequest{})
	if err != nil {
		return nil, convError(err)
	}
	entryChan := make(chan dinkur.StreamedEntry)
	go func() {
		for {
			res, err := stream.Recv()
			if err != nil {
				if err != io.EOF {
					log.Error().
						WithError(convError(err)).
						Message("Error when streaming entries. Closing stream.")
				}
				close(entryChan)
				return
			}
			if res == nil {
				continue
			}
			const logWarnMsg = "Error when streaming entries. Ignoring message."
			entry, err := convEntryPtr(res.Entry)
			if err != nil {
				log.Warn().WithError(convError(err)).
					Message(logWarnMsg)
				continue
			}
			if entry == nil {
				log.Warn().WithError(ErrUnexpectedNilEntry).
					Message(logWarnMsg)
				continue
			}
			entryChan <- dinkur.StreamedEntry{
				Entry:  *entry,
				Event: convEvent(res.Event),
			}
		}
	}()
	return entryChan, nil
}
