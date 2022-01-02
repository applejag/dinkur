// Dinkur the task time tracking utility.
// <https://github.com/dinkur/dinkur>
//
// SPDX-FileCopyrightText: 2021 Kalle Fagerberg
// SPDX-License-Identifier: GPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify it under
// the terms of the GNU General Public License as published by the
// Free Software Foundation, either version 3 of the License, or (at your option)
// any later version.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
// FOR A PARTICULAR PURPOSE.  See the GNU General Public License for more
// details.
//
// You should have received a copy of the GNU General Public License along with
// this program.  If not, see <http://www.gnu.org/licenses/>.

package dinkurafk

import (
	"errors"
	"sync"
)

// Errors specific for the listener and subscriptions.
var (
	ErrUnsupportedSubType        = errors.New("unsupported subscription type")
	ErrAlreadyUnsubscribed       = errors.New("already unsubscribed")
	ErrSubscriptionNotInitalized = errors.New("subscription is not initialized")
)

// Started contains event data for when user has gone AFK.
type Started struct {
}

// Stopped contains event data for when user is no longer AFK (after being AFK).
type Stopped struct {
}

// SubStarted is a subscription of events for when user has gone AFK.
type SubStarted interface {
	Started() <-chan Started
}

// SubStopped is a subscription of events for when user is no longer AFK
// (after being AFK).
type SubStopped interface {
	Stopped() <-chan Stopped
}

type subStarted struct {
	id uint
	c  chan Started
}

func (sub *subStarted) pubWaitGroup(s Started, wg *sync.WaitGroup) {
	sub.c <- s
	wg.Done()
}

func (sub subStarted) Started() <-chan Started {
	return sub.c
}

type subStopped struct {
	id uint
	c  chan Stopped
}

func (sub *subStopped) pubWaitGroup(s Stopped, wg *sync.WaitGroup) {
	sub.c <- s
	wg.Done()
}

func (sub subStopped) Stopped() <-chan Stopped {
	return sub.c
}

// NewObserver returns a new AFK-started and AFK-stopped observer.
func NewObserver() Observer {
	return &observer{}
}

type observer struct {
	obsStarted
	obsStopped
}

// Observer lets you publish and subscribe to both AFK-started and AFK-stopped
// events.
type Observer interface {
	ObserverStarted
	ObserverStopped
}

// ObserverStarted lets you publish and subscribe to AFK-started events.
type ObserverStarted interface {
	// PubStartedWait publishes an AFK-started event and waits until all
	// subscriptions has received their events.
	PubStartedWait(Started)
	SubStarted() SubStarted
	UnsubStarted(SubStarted) error
}

// ObserverStopped lets you publish and subscribe to AFK-stopped events.
type ObserverStopped interface {
	// PubStoppedWait publishes an AFK-stopped event and waits until all
	// subscriptions has received their events.
	PubStoppedWait(Stopped)
	SubStopped() SubStopped
	UnsubStopped(SubStopped) error
}

type obsStarted struct {
	nextID uint
	subs   []subStarted
	mutex  sync.RWMutex
}

func (o *obsStarted) SubStarted() SubStarted {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.nextID++
	sub := subStarted{
		id: o.nextID,
		c:  make(chan Started),
	}
	o.subs = append(o.subs, sub)
	return sub
}

func (o *obsStarted) PubStartedWait(s Started) {
	var wg sync.WaitGroup
	wg.Add(len(o.subs))
	o.mutex.RLock()
	for _, sub := range o.subs {
		go sub.pubWaitGroup(s, &wg)
	}
	o.mutex.RUnlock()
	wg.Wait()
}

func (o *obsStarted) UnsubStarted(sub SubStarted) error {
	s, ok := sub.(subStarted)
	if !ok {
		return ErrUnsupportedSubType
	}
	if s.id == 0 {
		return ErrSubscriptionNotInitalized
	}
	idx := o.subStartedIndex(s.id)
	if idx == -1 {
		return ErrAlreadyUnsubscribed
	}
	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.subs = append(o.subs[idx:], o.subs[idx+1:]...)
	return nil
}

func (o *obsStarted) subStartedIndex(id uint) int {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	for i, sub := range o.subs {
		if sub.id == id {
			return i
		}
	}
	return -1
}

type obsStopped struct {
	nextID uint
	subs   []subStopped
	mutex  sync.RWMutex
}

func (o *obsStopped) SubStopped() SubStopped {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.nextID++
	sub := subStopped{
		id: o.nextID,
		c:  make(chan Stopped),
	}
	o.subs = append(o.subs, sub)
	return sub
}

func (o *obsStopped) PubStoppedWait(s Stopped) {
	var wg sync.WaitGroup
	wg.Add(len(o.subs))
	o.mutex.RLock()
	for _, sub := range o.subs {
		go sub.pubWaitGroup(s, &wg)
	}
	o.mutex.RUnlock()
	wg.Wait()
}

func (o *obsStopped) UnsubStopped(sub SubStopped) error {
	s, ok := sub.(subStopped)
	if !ok {
		return ErrUnsupportedSubType
	}
	if s.id == 0 {
		return ErrSubscriptionNotInitalized
	}
	idx := o.subStoppedIndex(s.id)
	if idx == -1 {
		return ErrAlreadyUnsubscribed
	}
	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.subs = append(o.subs[idx:], o.subs[idx+1:]...)
	return nil
}

func (o *obsStopped) subStoppedIndex(id uint) int {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	for i, sub := range o.subs {
		if sub.id == id {
			return i
		}
	}
	return -1
}
