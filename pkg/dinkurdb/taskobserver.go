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

type taskEvent struct {
	dbTask Task
	event  dinkur.EventType
}

type taskObserver struct {
	subs  []chan taskEvent
	mutex sync.RWMutex
}

func (o *taskObserver) pubTask(ev taskEvent) {
	o.mutex.RLock()
	for _, sub := range o.subs {
		go func(ev taskEvent, sub chan taskEvent) {
			select {
			case sub <- ev:
			case <-time.After(10 * time.Second):
				log.Warn().
					WithUint("id", ev.dbTask.ID).
					WithString("name", ev.dbTask.Name).
					WithStringer("event", ev.event).
					Message("Timed out sending task event.")
			}
		}(ev, sub)
	}
	o.mutex.RUnlock()
}

func (o *taskObserver) subTasks() <-chan taskEvent {
	o.mutex.Lock()
	sub := make(chan taskEvent)
	o.subs = append(o.subs, sub)
	o.mutex.Unlock()
	return sub
}

func (o *taskObserver) unsubTasks(sub <-chan taskEvent) error {
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

func (o *taskObserver) unsubAllTasks() error {
	o.mutex.Lock()
	o.subs = nil
	o.mutex.Unlock()
	return nil
}

func (o *taskObserver) subIndex(sub <-chan taskEvent) int {
	for i, ch := range o.subs {
		if ch == sub {
			return i
		}
	}
	return -1
}
