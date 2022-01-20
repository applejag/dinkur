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

package dinkurdb

import (
	"errors"
	"sync"
	"time"

	"github.com/dinkur/dinkur/pkg/dinkur"
)

// Errors specific for the listener and subscriptions.
var (
	errAlreadyUnsubscribed       = errors.New("already unsubscribed")
	errSubscriptionNotInitalized = errors.New("subscription is not initialized")
)

type entryEvent struct {
	dbEntry Entry
	event  dinkur.EventType
}

type entryObserver struct {
	subs  []chan entryEvent
	mutex sync.RWMutex
}

func (o *entryObserver) pubEntry(ev entryEvent) {
	o.mutex.RLock()
	for _, sub := range o.subs {
		go func(ev entryEvent, sub chan entryEvent) {
			select {
			case sub <- ev:
			case <-time.After(10 * time.Second):
				log.Warn().
					WithUint("id", ev.dbEntry.ID).
					WithString("name", ev.dbEntry.Name).
					WithStringer("event", ev.event).
					Message("Timed out sending entry event.")
			}
		}(ev, sub)
	}
	o.mutex.RUnlock()
}

func (o *entryObserver) subEntries() <-chan entryEvent {
	o.mutex.Lock()
	sub := make(chan entryEvent)
	o.subs = append(o.subs, sub)
	o.mutex.Unlock()
	return sub
}

func (o *entryObserver) unsubEntries(sub <-chan entryEvent) error {
	if sub == nil {
		return errSubscriptionNotInitalized
	}
	o.mutex.Lock()
	defer o.mutex.Unlock()
	idx := o.subIndex(sub)
	if idx == -1 {
		return errAlreadyUnsubscribed
	}
	o.subs = append(o.subs[:idx], o.subs[idx+1:]...)
	return nil
}

func (o *entryObserver) unsubAllEntries() error {
	o.mutex.Lock()
	o.subs = nil
	o.mutex.Unlock()
	return nil
}

func (o *entryObserver) subIndex(sub <-chan entryEvent) int {
	for i, ch := range o.subs {
		if ch == sub {
			return i
		}
	}
	return -1
}
