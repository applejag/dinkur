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
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/dinkur/dinkur/pkg/dinkur"
	"github.com/dinkur/dinkur/pkg/timeutil"
)

func (c *client) GetActiveEntry(ctx context.Context) (*dinkur.Entry, error) {
	dbEntry, err := c.withContext(ctx).activeDBEntry()
	if err != nil {
		return nil, err
	}
	return convEntryPtr(dbEntry), nil
}

func (c *client) activeDBEntry() (*Entry, error) {
	if err := c.assertConnected(); err != nil {
		return nil, err
	}
	var dbEntry Entry
	err := c.db.Where(Entry{End: nil}, entryFieldEnd).First(&dbEntry).Error
	if err != nil {
		return nil, nilNotFoundError(err)
	}
	return &dbEntry, nil
}

func (c *client) GetEntry(ctx context.Context, id uint) (dinkur.Entry, error) {
	dbEntry, err := c.withContext(ctx).getDBEntry(id)
	if err != nil {
		return dinkur.Entry{}, err
	}
	return convEntry(dbEntry), nil
}

func (c *client) getDBEntry(id uint) (Entry, error) {
	if err := c.assertConnected(); err != nil {
		return Entry{}, err
	}
	var dbEntry Entry
	err := c.db.First(&dbEntry, id).Error
	if err != nil {
		return Entry{}, err
	}
	return dbEntry, nil
}

var (
	entrySQLBetweenStart = fmt.Sprintf(
		"((%[1]s >= @start) OR "+
			"(%[2]s IS NOT NULL AND %[1]s >= @start) OR "+
			"(%[2]s IS NULL AND CURRENT_TIMESTAMP >= @start))",
		entryColumnStart, entryColumnEnd,
	)

	entrySQLBetweenEnd = fmt.Sprintf(
		"((%[2]s <= @end) OR "+
			"(%[2]s IS NOT NULL AND %[2]s <= @end) OR "+
			"(%[2]s IS NULL AND CURRENT_TIMESTAMP <= @end))",
		entryColumnStart, entryColumnEnd,
	)

	entrySQLBetween = fmt.Sprintf(
		"((%[1]s BETWEEN @start AND @end) OR "+
			"(%[2]s IS NOT NULL AND %[2]s BETWEEN @start AND @end) OR "+
			"(%[2]s IS NULL AND CURRENT_TIMESTAMP BETWEEN @start AND @end))",
		entryColumnStart, entryColumnEnd,
	)
)

func (c *client) GetEntryList(ctx context.Context, search dinkur.SearchEntry) ([]dinkur.Entry, error) {
	dbEntries, err := c.withContext(ctx).listDBEntries(search)
	if err != nil {
		return nil, err
	}
	return convEntrySlice(dbEntries), nil
}

func (c *client) listDBEntries(search dinkur.SearchEntry) ([]Entry, error) {
	if err := c.assertConnected(); err != nil {
		return nil, err
	}
	span := search.Shorthand.Span(time.Now())
	if search.Start == nil {
		search.Start = span.Start
	}
	if search.End == nil {
		search.End = span.End
	}
	if search.Limit > math.MaxInt {
		return nil, dinkur.ErrLimitTooLarge
	}
	var dbEntries []Entry
	q := c.db.Model(&Entry{}).
		Order(entryColumnStart + " DESC").
		Limit(int(search.Limit))
	switch {
	case search.Start != nil && search.End != nil:
		// adding/subtracting 1s to resolve rounding issues, as Sqlite's
		// smallest time unit is a second.
		start := (*search.Start).UTC().Add(-time.Second)
		end := (*search.End).UTC().Add(time.Second)
		q = q.Where(entrySQLBetween, sql.Named("start", start), sql.Named("end", end))
	case search.Start != nil:
		start := (*search.Start).UTC().Add(-time.Second)
		q = q.Where(entrySQLBetweenStart, sql.Named("start", start))
	case search.End != nil:
		end := (*search.End).UTC().Add(time.Second)
		q = q.Where(entrySQLBetweenEnd, sql.Named("end", end))
	}
	if search.NameFuzzy != "" {
		if search.NameHighlightStart != "" || search.NameHighlightEnd != "" {
			q = q.Joins("INNER JOIN entries_idx ON entries.id = entries_idx.rowid").
				Select(
					"id, created_at, updated_at, highlight(entries_idx, 0, ?, ?) AS name, start, end",
					search.NameHighlightStart, search.NameHighlightEnd).
				Where(entryFTS5ColumnName+" MATCH ?", search.NameFuzzy)
		} else {
			subQ := c.db.Model(&EntryFTS5{}).
				Select(entryFTS5ColumnRowID).
				Where(entryFTS5ColumnName+" MATCH ?", search.NameFuzzy)
			q = q.Where(entryColumnID+" IN (?)", subQ)
		}
	}
	if err := q.Find(&dbEntries).Error; err != nil {
		return nil, err
	}
	// we sorted in descending order to get the last entries.
	// fix this by reversing "again"
	reverseEntrySlice(dbEntries)
	return dbEntries, nil
}

func (c *client) UpdateEntry(ctx context.Context, edit dinkur.EditEntry) (dinkur.UpdatedEntry, error) {
	if err := c.assertConnected(); err != nil {
		return dinkur.UpdatedEntry{}, err
	}
	update, err := c.withContext(ctx).editDBEntry(edit)
	if err != nil {
		return dinkur.UpdatedEntry{}, err
	}
	c.entryObs.pubEntry(entryEvent{
		dbEntry: update.after,
		event:  dinkur.EventUpdated,
	})
	return dinkur.UpdatedEntry{
		Before: convEntry(update.before),
		After:  convEntry(update.after),
	}, nil
}

type updatedDBEntry struct {
	before Entry
	after  Entry
}

func (c *client) editDBEntry(edit dinkur.EditEntry) (updatedDBEntry, error) {
	if edit.Name != nil && *edit.Name == "" {
		return updatedDBEntry{}, dinkur.ErrEntryNameEmpty
	}
	if edit.Start != nil && edit.End != nil && edit.Start.After(*edit.End) {
		return updatedDBEntry{}, dinkur.ErrEntryEndBeforeStart
	}
	var update updatedDBEntry
	err := c.transaction(func(tx *client) (tranErr error) {
		update, tranErr = tx.editDBEntryNoTran(edit)
		return
	})
	return update, err
}

func (c *client) editDBEntryNoTran(edit dinkur.EditEntry) (updatedDBEntry, error) {
	dbEntry, err := c.getDBEntryToEditNoTran(edit.IDOrZero)
	if err != nil {
		if errors.Is(err, dinkur.ErrNotFound) {
			return updatedDBEntry{}, fmt.Errorf("no entry to edit, failed finding latest entry: %w", err)
		}
		return updatedDBEntry{}, fmt.Errorf("get entry to edit: %w", err)
	}
	startAfterTime, err := c.getTimeToStartAfterOrNow(edit.StartAfterIDOrZero, edit.StartAfterLast)
	if err != nil {
		return updatedDBEntry{}, err
	}
	if startAfterTime != nil {
		edit.Start = startAfterTime
	}
	endBeforeTime, err := c.getTimeToEndBefore(edit.EndBeforeIDOrZero)
	if err != nil {
		return updatedDBEntry{}, err
	}
	if endBeforeTime != nil {
		edit.End = endBeforeTime
	}
	var anyEdit bool
	entryBeforeEdit := dbEntry
	if edit.Name != nil {
		if edit.AppendName {
			dbEntry.Name = fmt.Sprint(dbEntry.Name, " ", *edit.Name)
		} else {
			dbEntry.Name = *edit.Name
		}
		anyEdit = true
	}
	if edit.Start != nil {
		dbEntry.Start = *edit.Start
		anyEdit = true
	}
	if edit.End != nil {
		dbEntry.End = edit.End
		anyEdit = true
	}
	if dbEntry.Elapsed() < 0 {
		return updatedDBEntry{}, dinkur.ErrEntryEndBeforeStart
	}
	if anyEdit {
		if err := c.db.Save(&dbEntry).Error; err != nil {
			return updatedDBEntry{}, fmt.Errorf("save updated entry: %w", err)
		}
	}
	return updatedDBEntry{
		before: entryBeforeEdit,
		after:  dbEntry,
	}, nil
}

func (c *client) getDBEntryToStartAfter(idOrZero uint, lastEntry bool) (*Entry, error) {
	if idOrZero != 0 {
		startAfter, err := c.getDBEntry(idOrZero)
		if err != nil {
			return nil, fmt.Errorf("get entry by ID to start after: %w", err)
		}
		return &startAfter, nil
	} else if lastEntry {
		lastEntries, err := c.listDBEntries(dinkur.SearchEntry{
			Shorthand: timeutil.TimeSpanNone,
			Limit:     1,
		})
		if err != nil {
			return nil, fmt.Errorf("get last entry to start after: %w", err)
		}
		if len(lastEntries) == 0 {
			return nil, fmt.Errorf("get last entry to start after: %w", dinkur.ErrNotFound)
		}
		return &lastEntries[0], nil
	}
	return nil, nil
}

func (c *client) getTimeToStartAfterOrNow(idOrZero uint, lastEntry bool) (*time.Time, error) {
	startAfter, err := c.getDBEntryToStartAfter(idOrZero, lastEntry)
	if err != nil {
		return nil, err
	}
	if startAfter == nil {
		return nil, nil
	}
	if startAfter.End == nil {
		now := time.Now()
		return &now, nil
	}
	return startAfter.End, nil
}

func (c *client) getDBEntryToEndBefore(idOrZero uint) (*Entry, error) {
	if idOrZero == 0 {
		return nil, nil
	}
	endBefore, err := c.getDBEntry(idOrZero)
	if err != nil {
		return nil, fmt.Errorf("get entry by ID to end before: %w", err)
	}
	return &endBefore, nil
}

func (c *client) getTimeToEndBefore(idOrZero uint) (*time.Time, error) {
	endBefore, err := c.getDBEntryToEndBefore(idOrZero)
	if err != nil {
		return nil, err
	}
	if endBefore == nil {
		return nil, nil
	}
	return &endBefore.Start, nil
}

func (c *client) getDBEntryToEditNoTran(idOrZero uint) (Entry, error) {
	if idOrZero != 0 {
		dbEntryByID, err := c.getDBEntry(idOrZero)
		if err != nil {
			return Entry{}, fmt.Errorf("get entry by ID: %d: %w", idOrZero, err)
		}
		return dbEntryByID, nil
	}
	activeDBEntry, err := c.activeDBEntry()
	if err != nil {
		return Entry{}, fmt.Errorf("get active entry: %w", err)
	}
	if activeDBEntry != nil {
		return *activeDBEntry, nil
	}
	now := time.Now()
	dbEntries, err := c.listDBEntries(dinkur.SearchEntry{
		Limit: 1,
		End:   &now,
	})
	if err != nil {
		return Entry{}, fmt.Errorf("list latest 1 entry: %w", err)
	}
	if len(dbEntries) == 0 {
		return Entry{}, dinkur.ErrNotFound
	}
	return dbEntries[0], nil
}

func (c *client) DeleteEntry(ctx context.Context, id uint) (dinkur.Entry, error) {
	if err := c.assertConnected(); err != nil {
		return dinkur.Entry{}, err
	}
	dbEntry, err := c.withContext(ctx).deleteDBEntry(id)
	if err != nil {
		return dinkur.Entry{}, err
	}
	c.entryObs.pubEntry(entryEvent{
		dbEntry: dbEntry,
		event:  dinkur.EventDeleted,
	})
	return convEntry(dbEntry), err
}

func (c *client) deleteDBEntry(id uint) (Entry, error) {
	var dbEntry Entry
	err := c.transaction(func(tx *client) (tranErr error) {
		dbEntry, tranErr = tx.deleteDBEntryNoTran(id)
		return
	})
	return dbEntry, err
}

func (c *client) deleteDBEntryNoTran(id uint) (Entry, error) {
	dbEntry, err := c.getDBEntry(id)
	if err != nil {
		return Entry{}, fmt.Errorf("get entry to delete: %w", err)
	}
	if err := c.db.Delete(&Entry{}, id).Error; err != nil {
		return Entry{}, fmt.Errorf("delete entry: %w", err)
	}
	return dbEntry, nil
}

func (c *client) CreateEntry(ctx context.Context, entry dinkur.NewEntry) (dinkur.StartedEntry, error) {
	if err := c.assertConnected(); err != nil {
		return dinkur.StartedEntry{}, err
	}
	if entry.Name == "" {
		return dinkur.StartedEntry{}, dinkur.ErrEntryNameEmpty
	}
	var start time.Time
	if entry.Start != nil {
		start = *entry.Start
	} else {
		start = time.Now()
	}
	if entry.End != nil && entry.End.Before(start) {
		return dinkur.StartedEntry{}, dinkur.ErrEntryEndBeforeStart
	}
	newEntry := newEntry{
		Entry: Entry{
			Name:  entry.Name,
			Start: start.UTC(),
			End:   timePtrUTC(entry.End),
		},
		startAfterIDOrZero: entry.StartAfterIDOrZero,
		endBeforeIDOrZero:  entry.EndBeforeIDOrZero,
		startAfterLast:     entry.StartAfterLast,
	}
	startedEntry, err := c.withContext(ctx).startDBEntry(newEntry)
	if err != nil {
		return dinkur.StartedEntry{}, err
	}
	if startedEntry.stopped != nil {
		c.entryObs.pubEntry(entryEvent{
			dbEntry: *startedEntry.stopped,
			event:  dinkur.EventUpdated,
		})
	}
	c.entryObs.pubEntry(entryEvent{
		dbEntry: startedEntry.started,
		event:  dinkur.EventCreated,
	})
	return dinkur.StartedEntry{
		Started: convEntry(startedEntry.started),
		Stopped: convEntryPtr(startedEntry.stopped),
	}, nil
}

type startedDBEntry struct {
	started Entry
	stopped *Entry
}

type newEntry struct {
	Entry
	startAfterIDOrZero uint
	endBeforeIDOrZero  uint
	startAfterLast     bool
}

func (c *client) startDBEntry(newEntry newEntry) (startedDBEntry, error) {
	var startedEntry startedDBEntry
	err := c.transaction(func(tx *client) (tranErr error) {
		startedEntry, tranErr = tx.startDBEntryNoTran(newEntry)
		return
	})
	return startedEntry, err
}

func (c *client) startDBEntryNoTran(newEntry newEntry) (startedDBEntry, error) {
	startAfterTime, err := c.getTimeToStartAfterOrNow(newEntry.startAfterIDOrZero, newEntry.startAfterLast)
	if err != nil {
		return startedDBEntry{}, err
	}
	if startAfterTime != nil {
		newEntry.Start = *startAfterTime
	}
	endBeforeTime, err := c.getTimeToEndBefore(newEntry.endBeforeIDOrZero)
	if err != nil {
		return startedDBEntry{}, err
	}
	if endBeforeTime != nil {
		newEntry.End = endBeforeTime
	}
	previousDBEntry, err := c.stopActiveDBEntryNoTran(newEntry.Start)
	if err != nil {
		return startedDBEntry{}, fmt.Errorf("stop previously active entry: %w", err)
	}
	err = c.db.Create(&newEntry.Entry).Error
	if err != nil {
		return startedDBEntry{}, fmt.Errorf("create new active entry: %w", err)
	}
	return startedDBEntry{
		stopped: previousDBEntry,
		started: newEntry.Entry,
	}, nil
}

func (c *client) StopActiveEntry(ctx context.Context, endTime time.Time) (*dinkur.Entry, error) {
	if err := c.assertConnected(); err != nil {
		return nil, err
	}
	dbEntry, err := c.withContext(ctx).stopActiveDBEntry(endTime)
	if err != nil {
		return nil, err
	}
	if err == nil && dbEntry != nil {
		c.entryObs.pubEntry(entryEvent{
			dbEntry: *dbEntry,
			event:  dinkur.EventUpdated,
		})
	}
	return convEntryPtr(dbEntry), nil
}

func (c *client) stopActiveDBEntry(endTime time.Time) (*Entry, error) {
	var activeDBEntry *Entry
	err := c.transaction(func(tx *client) (tranErr error) {
		activeDBEntry, tranErr = tx.stopActiveDBEntryNoTran(endTime)
		return
	})
	return activeDBEntry, err
}

func (c *client) stopActiveDBEntryNoTran(endTime time.Time) (*Entry, error) {
	var entries []Entry
	if err := c.db.Where(&Entry{End: nil}, entryFieldEnd).Find(&entries).Error; err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, nil
	}
	for i, entry := range entries {
		if endTime.Before(entry.Start) {
			return nil, dinkur.ErrEntryEndBeforeStart
		}
		entries[i].End = &endTime
	}
	err := c.db.Model(&Entry{}).
		Where(&Entry{End: nil}, entryFieldEnd).
		Update(entryFieldEnd, endTime).
		Error
	if err != nil {
		return nil, err
	}
	return &entries[0], nil
}

func (c *client) StreamEntry(ctx context.Context) (<-chan dinkur.StreamedEntry, error) {
	ch := make(chan dinkur.StreamedEntry)
	go func() {
		done := ctx.Done()
		dbEntryChan := c.entryObs.subEntries()
		defer close(ch)
		defer func() {
			if err := c.entryObs.unsubEntries(dbEntryChan); err != nil {
				log.Warn().WithError(err).Message("Failed to unsub entry.")
			}
		}()
		for {
			select {
			case ev, ok := <-dbEntryChan:
				if !ok {
					return
				}
				ch <- dinkur.StreamedEntry{
					Entry:  convEntry(ev.dbEntry),
					Event: ev.event,
				}
			case <-done:
				return
			}
		}
	}()
	return ch, nil
}

func reverseEntrySlice(slice []Entry) {
	for i, j := 0, len(slice)-1; i < j; i, j = i+1, j-1 {
		slice[i], slice[j] = slice[j], slice[i]
	}
}
