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
	"time"

	"github.com/dinkur/dinkur/pkg/dbmodel"
	"github.com/dinkur/dinkur/pkg/dinkur"
	"github.com/dinkur/dinkur/pkg/fromdb"
	"gopkg.in/typ.v4"
)

func (c *client) StreamStatus(ctx context.Context) (<-chan dinkur.StreamedStatus, error) {
	if err := c.assertConnected(); err != nil {
		return nil, err
	}
	ch := make(chan dinkur.StreamedStatus)
	go func() {
		done := ctx.Done()
		dbStatusChan := c.statusObs.Sub()
		defer close(ch)
		defer func() {
			if err := c.statusObs.Unsub(dbStatusChan); err != nil {
				log.Warn().WithError(err).Message("Failed to unsub status.")
			}
		}()
		for {
			select {
			case ev, ok := <-dbStatusChan:
				if !ok {
					return
				}
				ch <- dinkur.StreamedStatus{
					Status: fromdb.Status(ev.dbStatus),
				}
			case <-done:
				return
			}
		}
	}()
	return ch, nil

}

func (c *client) GetStatus(ctx context.Context) (dinkur.Status, error) {
	if err := c.assertConnected(); err != nil {
		return dinkur.Status{}, err
	}
	dbStatus, err := c.withContext(ctx).getDBStatusAtom()
	if err != nil {
		return dinkur.Status{}, err
	}
	return fromdb.Status(dbStatus), nil
}

func (c *client) getDBStatusAtom() (dbmodel.Status, error) {
	var dbStatus dbmodel.Status
	if err := c.db.Find(&dbStatus).Error; err != nil {
		return dbmodel.Status{}, err
	}
	return dbStatus, nil
}

func (c *client) SetStatus(ctx context.Context, edit dinkur.EditStatus) error {
	if err := c.assertConnected(); err != nil {
		return err
	}
	return c.withContext(ctx).setStatus(edit)
}

func (c *client) setStatus(edit dinkur.EditStatus) error {
	return c.transaction(func(tx *client) error {
		return tx.setStatusNoTran(edit)
	})
}

func (c *client) setStatusNoTran(edit dinkur.EditStatus) error {
	dbStatus, err := c.getDBStatusAtom()
	if err != nil {
		return err
	}
	var changed bool
	if updateTimePtrUTC(&dbStatus.AFKSince, edit.AFKSince) {
		changed = true
	}
	if updateTimePtrUTC(&dbStatus.BackSince, edit.BackSince) {
		changed = true
	}
	if !changed {
		return nil
	}
	if err := c.db.Save(&dbStatus).Error; err != nil {
		return err
	}
	c.statusObs.PubWait(statusEvent{dbStatus})
	return nil
}

func updateTimePtrUTC(ptr **time.Time, newValue *time.Time) bool {
	if (*ptr == nil) != (newValue == nil) || newValue != nil {
		if newValue != nil {
			*ptr = typ.Ref(newValue.UTC())
		} else {
			*ptr = nil
		}
		return true
	}
	return false
}
