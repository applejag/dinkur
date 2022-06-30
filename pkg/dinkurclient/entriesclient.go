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

	"github.com/dinkur/dinkur/pkg/conv"
	"github.com/dinkur/dinkur/pkg/dinkur"
	"github.com/dinkur/dinkur/pkg/fromgrpc"
	"github.com/dinkur/dinkur/pkg/togrpc"
	"google.golang.org/grpc"

	dinkurapiv1 "github.com/dinkur/dinkur/api/dinkurapi/v1"
)

type grpcFunc[Req, Res any] func(context.Context, Req, ...grpc.CallOption) (Res, error)

func invoke[Req any, Res ~*ResPtr, ResPtr any](ctx context.Context, c *client, f grpcFunc[Req, Res], req Req) (Res, error) {
	if err := c.assertConnected(); err != nil {
		return nil, err
	}
	res, err := f(ctx, req)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, ErrResponseIsNil
	}
	return res, nil
}

func (c *client) GetEntry(ctx context.Context, id uint) (dinkur.Entry, error) {
	res, err := invoke(ctx, c, c.entryer.GetEntry, &dinkurapiv1.GetEntryRequest{
		Id: uint64(id),
	})
	if err != nil {
		return dinkur.Entry{}, convError(err)
	}
	entry, err := fromgrpc.EntryPtrNoNil(res.Entry)
	if err != nil {
		return dinkur.Entry{}, convError(err)
	}
	return entry, nil
}

func (c *client) GetEntryList(ctx context.Context, search dinkur.SearchEntry) ([]dinkur.Entry, error) {
	res, err := invoke(ctx, c, c.entryer.GetEntryList, &dinkurapiv1.GetEntryListRequest{
		Start:              togrpc.TimestampPtr(search.Start),
		End:                togrpc.TimestampPtr(search.End),
		Limit:              uint64(search.Limit),
		Shorthand:          togrpc.Shorthand(search.Shorthand),
		NameFuzzy:          search.NameFuzzy,
		NameHighlightStart: search.NameHighlightStart,
		NameHighlightEnd:   search.NameHighlightEnd,
	})
	if err != nil {
		return nil, convError(err)
	}
	entries, err := fromgrpc.EntrySlice(res.Entries)
	if err != nil {
		return nil, convError(err)
	}
	return entries, nil
}

func (c *client) UpdateEntry(ctx context.Context, edit dinkur.EditEntry) (dinkur.UpdatedEntry, error) {
	res, err := invoke(ctx, c, c.entryer.UpdateEntry, &dinkurapiv1.UpdateEntryRequest{
		IdOrZero:           uint64(edit.IDOrZero),
		Name:               conv.DerefOrZero(edit.Name),
		Start:              togrpc.TimestampPtr(edit.Start),
		End:                togrpc.TimestampPtr(edit.End),
		Relative:           edit.Relative,
		AppendName:         edit.AppendName,
		StartAfterIdOrZero: uint64(edit.StartAfterIDOrZero),
		EndBeforeIdOrZero:  uint64(edit.EndBeforeIDOrZero),
		StartAfterLast:     edit.StartAfterLast,
	})
	if err != nil {
		return dinkur.UpdatedEntry{}, convError(err)
	}
	entryBefore, err := fromgrpc.EntryPtrNoNil(res.Before)
	if err != nil {
		return dinkur.UpdatedEntry{}, fmt.Errorf("entry before: %w", convError(err))
	}
	entryAfter, err := fromgrpc.EntryPtrNoNil(res.After)
	if err != nil {
		return dinkur.UpdatedEntry{}, fmt.Errorf("entry after: %w", convError(err))
	}
	return dinkur.UpdatedEntry{
		Before: entryBefore,
		After:  entryAfter,
	}, nil
}

func (c *client) DeleteEntry(ctx context.Context, id uint) (dinkur.Entry, error) {
	res, err := invoke(ctx, c, c.entryer.DeleteEntry, &dinkurapiv1.DeleteEntryRequest{
		Id: uint64(id),
	})
	if err != nil {
		return dinkur.Entry{}, convError(err)
	}
	entry, err := fromgrpc.EntryPtrNoNil(res.DeletedEntry)
	if err != nil {
		return dinkur.Entry{}, convError(err)
	}
	return entry, nil
}

func (c *client) CreateEntry(ctx context.Context, entry dinkur.NewEntry) (dinkur.StartedEntry, error) {
	res, err := invoke(ctx, c, c.entryer.CreateEntry, &dinkurapiv1.CreateEntryRequest{
		Name:               entry.Name,
		Start:              togrpc.TimestampPtr(entry.Start),
		End:                togrpc.TimestampPtr(entry.End),
		StartAfterIdOrZero: uint64(entry.StartAfterIDOrZero),
		EndBeforeIdOrZero:  uint64(entry.EndBeforeIDOrZero),
		StartAfterLast:     entry.StartAfterLast,
	})
	if err != nil {
		return dinkur.StartedEntry{}, convError(err)
	}
	prevEntry, err := fromgrpc.EntryPtr(res.PreviouslyActiveEntry)
	if err != nil {
		return dinkur.StartedEntry{}, fmt.Errorf("stopped entry: %w", convError(err))
	}
	newEntry, err := fromgrpc.EntryPtrNoNil(res.CreatedEntry)
	if err != nil {
		return dinkur.StartedEntry{}, fmt.Errorf("created entry: %w", convError(err))
	}
	return dinkur.StartedEntry{
		Stopped: prevEntry,
		Started: newEntry,
	}, nil
}

func (c *client) GetActiveEntry(ctx context.Context) (*dinkur.Entry, error) {
	res, err := invoke(ctx, c, c.entryer.GetActiveEntry, &dinkurapiv1.GetActiveEntryRequest{})
	if err != nil {
		return nil, convError(err)
	}
	entry, err := fromgrpc.EntryPtr(res.ActiveEntry)
	if err != nil {
		return nil, convError(err)
	}
	return entry, nil
}

func (c *client) StopActiveEntry(ctx context.Context, endTime time.Time) (*dinkur.Entry, error) {
	res, err := invoke(ctx, c, c.entryer.StopActiveEntry, &dinkurapiv1.StopActiveEntryRequest{
		End: togrpc.TimestampPtr(&endTime),
	})
	if err != nil {
		return nil, convError(err)
	}
	entry, err := fromgrpc.EntryPtr(res.StoppedEntry)
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
			entry, err := fromgrpc.EntryPtr(res.Entry)
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
				Entry: *entry,
				Event: fromgrpc.Event(res.Event),
			}
		}
	}()
	return entryChan, nil
}
