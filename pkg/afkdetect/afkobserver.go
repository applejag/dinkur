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

package afkdetect

import (
	"errors"
	"sync"
	"time"
)

// Errors specific for the listener and subscriptions.
var (
	ErrAlreadyUnsubscribed       = errors.New("already unsubscribed")
	ErrSubscriptionNotInitalized = errors.New("subscription is not initialized")
)

// Started contains event data for when user has gone AFK.
type Started struct{}

// Stopped contains event data for when user is no longer AFK (after being AFK).
type Stopped struct {
	AFKSince time.Time
}

// NewObserverStarted returns a new AFK-started events observer.
func NewObserverStarted() ObserverStarted {
	return &obsStarted{}
}

// NewObserverStopped returns a new AFK-stopped events observer.
func NewObserverStopped() ObserverStopped {
	return &obsStopped{}
}

// ObserverStarted lets you publish and subscribe to AFK-started events.
type ObserverStarted interface {
	// PubStartedWait publishes an AFK-started event and waits until all
	// subscriptions has received their events.
	PubStartedWait(Started)
	SubStarted() <-chan Started
	UnsubStarted(<-chan Started) error
	UnsubAllStarted() error
}

// ObserverStopped lets you publish and subscribe to AFK-stopped events.
type ObserverStopped interface {
	// PubStoppedWait publishes an AFK-stopped event and waits until all
	// subscriptions has received their events.
	PubStoppedWait(Stopped)
	SubStopped() <-chan Stopped
	UnsubStopped(<-chan Stopped) error
	UnsubAllStopped() error
}

type obsStarted struct {
	subs  []chan Started
	mutex sync.RWMutex
}

func (o *obsStarted) SubStarted() <-chan Started {
	o.mutex.Lock()
	sub := make(chan Started)
	o.subs = append(o.subs, sub)
	o.mutex.Unlock()
	return sub
}

func (o *obsStarted) PubStartedWait(s Started) {
	var wg sync.WaitGroup
	o.mutex.RLock()
	wg.Add(len(o.subs))
	for _, sub := range o.subs {
		go func(s Started, sub chan Started, wg *sync.WaitGroup) {
			sub <- s
			wg.Done()
		}(s, sub, &wg)
	}
	o.mutex.RUnlock()
	wg.Wait()
}

func (o *obsStarted) UnsubStarted(sub <-chan Started) error {
	if sub == nil {
		return ErrSubscriptionNotInitalized
	}
	o.mutex.Lock()
	defer o.mutex.Unlock()
	idx := o.subStartedIndex(sub)
	if idx == -1 {
		return ErrAlreadyUnsubscribed
	}
	o.subs = append(o.subs[idx:], o.subs[:idx+1]...)
	return nil
}

func (o *obsStarted) UnsubAllStarted() error {
	o.mutex.Lock()
	o.subs = nil
	o.mutex.Unlock()
	return nil
}

func (o *obsStarted) subStartedIndex(sub <-chan Started) int {
	for i, ch := range o.subs {
		if ch == sub {
			return i
		}
	}
	return -1
}

type obsStopped struct {
	chans []chan Stopped
	mutex sync.RWMutex
}

func (o *obsStopped) SubStopped() <-chan Stopped {
	o.mutex.Lock()
	ch := make(chan Stopped)
	o.chans = append(o.chans, ch)
	o.mutex.Unlock()
	return ch
}

func (o *obsStopped) PubStoppedWait(s Stopped) {
	var wg sync.WaitGroup
	o.mutex.RLock()
	wg.Add(len(o.chans))
	for _, ch := range o.chans {
		go func(s Stopped, ch chan Stopped, wg *sync.WaitGroup) {
			ch <- s
			wg.Done()
		}(s, ch, &wg)
	}
	o.mutex.RUnlock()
	wg.Wait()
}

func (o *obsStopped) UnsubStopped(sub <-chan Stopped) error {
	if sub == nil {
		return ErrSubscriptionNotInitalized
	}
	o.mutex.Lock()
	defer o.mutex.Unlock()
	idx := o.subStoppedIndex(sub)
	if idx == -1 {
		return ErrAlreadyUnsubscribed
	}
	o.chans = append(o.chans[idx:], o.chans[:idx+1]...)
	return nil
}

func (o *obsStopped) UnsubAllStopped() error {
	o.mutex.Lock()
	o.chans = nil
	o.mutex.Unlock()
	return nil
}

func (o *obsStopped) subStoppedIndex(sub <-chan Stopped) int {
	for i, ch := range o.chans {
		if ch == sub {
			return i
		}
	}
	return -1
}
