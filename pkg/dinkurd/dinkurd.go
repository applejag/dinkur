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

// Package dinkurd contains a Dinkur gRPC API server daemon implementation.
package dinkurd

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net"
	"sync"
	"time"

	dinkurapiv1 "github.com/dinkur/dinkur/api/dinkurapi/v1"
	"github.com/dinkur/dinkur/pkg/afkdetect"
	"github.com/dinkur/dinkur/pkg/dinkur"
	"github.com/dinkur/dinkur/pkg/dinkuralert"
	"github.com/dinkur/dinkur/pkg/timeutil"
	"github.com/iver-wharf/wharf-core/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Errors that are specific to the Dinkur gRPC server daemon.
var (
	ErrUintTooLarge   = fmt.Errorf("unsigned int value is too large, maximum: %d", uint64(math.MaxUint))
	ErrDaemonIsNil    = errors.New("daemon is nil")
	ErrRequestIsNil   = errors.New("grpc request was nil")
	ErrAlreadyServing = errors.New("daemon instance is already running")
)

var log = logger.NewScoped("daemon")

func convError(err error) error {
	switch {
	case status.Code(err) != codes.Unknown:
		return err
	case errors.Is(err, dinkur.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, ErrRequestIsNil),
		errors.Is(err, ErrUintTooLarge),
		errors.Is(err, dinkur.ErrLimitTooLarge),
		errors.Is(err, dinkur.ErrEntryEndBeforeStart),
		errors.Is(err, dinkur.ErrEntryNameEmpty):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, dinkur.ErrNotConnected),
		errors.Is(err, dinkur.ErrAlreadyConnected),
		errors.Is(err, dinkur.ErrClientIsNil):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return err
	}
}

func uint64ToUint(i uint64) (uint, error) {
	if i > math.MaxUint {
		return 0, ErrUintTooLarge
	}
	return uint(i), nil
}

func convUint64(i uint64) (uint, error) {
	if i > math.MaxUint {
		return 0, ErrUintTooLarge
	}
	return uint(i), nil
}

func convString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// Options for the daemon server.
type Options struct {
	// Host is the hostname to bind the server to.
	// Use 0.0.0.0 to allow any IP address.
	Host string
	// Port is the port the server will listen on.
	Port uint16
}

// DefaultOptions values are used for any zero values used when creating a new
// daemon instance.
var DefaultOptions = Options{
	Host: "localhost",
	Port: 59122,
}

// Daemon is the Dinkur daemon service interface.
type Daemon interface {
	// Serve starts the gRPC server and waits. The function does not return
	// unless the context is cancelled, or if there was an error.
	Serve(ctx context.Context) error
	// Close gracefully shuts down the daemon server.
	Close() error
}

// NewDaemon creates a new Daemon instance that relays all gRPC traffic to the
// given dinkur.Client. This daemon implementation does not perform any
// database communication nor has any persistence in of itself. This daemon
// must be paired with a dinkur.Client such as the dinkurdb client to talk to an
// Sqlite3 database file, or the dinkurclient client to act as a proxy.
//
// Both the global DefaultOptions and the opt parameter is used. The
// DefaultOptions values are only used for any zero valued fields in the
// opt parameter.
func NewDaemon(client dinkur.Client, opt Options) Daemon {
	if opt.Host == "" {
		opt.Host = DefaultOptions.Host
	}
	if opt.Port == 0 {
		opt.Port = DefaultOptions.Port
	}
	return &daemon{
		Options:     opt,
		client:      client,
		afkDetector: afkdetect.New(),
	}
}

type daemon struct {
	Options
	dinkurapiv1.UnimplementedEntriesServer
	dinkurapiv1.UnimplementedAlerterServer

	client     dinkur.Client
	grpcServer *grpc.Server
	listener   net.Listener

	afkDetector afkdetect.Detector
	closeMutex  sync.Mutex

	alertStore dinkuralert.Store
}

func (d *daemon) onEntryMutation() {
	d.alertStore.DeleteAFKAlert()
}

func (d *daemon) assertConnected() error {
	if d == nil {
		return ErrDaemonIsNil
	}
	if d.client == nil {
		return dinkur.ErrClientIsNil
	}
	return nil
}

func (d *daemon) Serve(ctx context.Context) error {
	if err := d.assertConnected(); err != nil {
		return err
	}
	if d.grpcServer != nil || d.listener != nil {
		return ErrAlreadyServing
	}
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", d.Host, d.Port))
	if err != nil {
		return fmt.Errorf("bind hostname and port: %w", err)
	}
	grpcServer := grpc.NewServer()
	d.listener = lis
	d.grpcServer = grpcServer
	defer d.Close()
	go func(ctx context.Context, d *daemon) {
		<-ctx.Done()
		d.Close()
	}(ctx, d)
	dinkurapiv1.RegisterEntriesServer(grpcServer, d)
	dinkurapiv1.RegisterAlerterServer(grpcServer, d)
	go d.listenForAFK(ctx)
	if err := d.afkDetector.StartDetecting(); err != nil {
		return fmt.Errorf("start afk detector: %w", err)
	}
	return grpcServer.Serve(lis)
}

func (d *daemon) Close() (finalErr error) {
	d.closeMutex.Lock()
	defer d.closeMutex.Unlock()
	if srv := d.grpcServer; srv != nil {
		srv.GracefulStop()
	}
	if lis := d.listener; lis != nil {
		if err := lis.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
			log.Error().WithError(err).Message("Closing grpc listener in Dinkur daemon.")
			finalErr = err
		}
	}
	d.grpcServer = nil
	d.listener = nil
	if err := d.afkDetector.StopDetecting(); err != nil {
		log.Error().WithError(err).Message("Stopping AFK detector in Dinkur daemon.")
		finalErr = err
	}
	return
}

func (d *daemon) listenForAFK(ctx context.Context) {
	log.Debug().Message("Listen for AFK events...")
	startedChan := d.afkDetector.StartedObs().Sub()
	stoppedChan := d.afkDetector.StoppedObs().Sub()
	defer d.afkDetector.StartedObs().Unsub(startedChan)
	defer d.afkDetector.StoppedObs().Unsub(stoppedChan)
	done := ctx.Done()
	for {
		select {
		case <-startedChan:
			entry, err := d.client.GetActiveEntry(ctx)
			if err != nil {
				log.Warn().WithError(err).
					Message("Failed to get active entry when issuing AFK alert.")
				continue
			}
			if entry == nil {
				continue
			}
			d.alertStore.SetAFK(*entry)
		case <-stoppedChan:
			d.alertStore.SetBackFromAFK()
		case <-done:
			return
		}
	}
}

func convEntryPtr(entry *dinkur.Entry) *dinkurapiv1.Entry {
	if entry == nil {
		return nil
	}
	return &dinkurapiv1.Entry{
		Id:      uint64(entry.ID),
		Created: convTime(entry.CreatedAt),
		Updated: convTime(entry.UpdatedAt),
		Name:    entry.Name,
		Start:   convTime(entry.Start),
		End:     convTimePtr(entry.End),
	}
}

func convEntrySlice(slice []dinkur.Entry) []*dinkurapiv1.Entry {
	entries := make([]*dinkurapiv1.Entry, len(slice))
	for i, t := range slice {
		entries[i] = convEntryPtr(&t)
	}
	return entries
}

func convTime(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}

func convTimePtr(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}

func convTimestampPtr(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t := ts.AsTime()
	return &t
}

func convTimestampOrNow(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Now()
	}
	t := ts.AsTime()
	return t
}

func convShorthand(s dinkurapiv1.GetEntryListRequest_Shorthand) timeutil.TimeSpanShorthand {
	switch s {
	case dinkurapiv1.GetEntryListRequest_SHORTHAND_PAST:
		return timeutil.TimeSpanPast
	case dinkurapiv1.GetEntryListRequest_SHORTHAND_FUTURE:
		return timeutil.TimeSpanFuture
	case dinkurapiv1.GetEntryListRequest_SHORTHAND_THIS_DAY:
		return timeutil.TimeSpanThisDay
	case dinkurapiv1.GetEntryListRequest_SHORTHAND_THIS_MON_TO_SUN:
		return timeutil.TimeSpanThisWeek
	case dinkurapiv1.GetEntryListRequest_SHORTHAND_PREV_DAY:
		return timeutil.TimeSpanPrevDay
	case dinkurapiv1.GetEntryListRequest_SHORTHAND_PREV_MON_TO_SUN:
		return timeutil.TimeSpanPrevWeek
	case dinkurapiv1.GetEntryListRequest_SHORTHAND_NEXT_DAY:
		return timeutil.TimeSpanNextDay
	case dinkurapiv1.GetEntryListRequest_SHORTHAND_NEXT_MON_TO_SUN:
		return timeutil.TimeSpanNextWeek
	default:
		return timeutil.TimeSpanNone
	}
}

func convAlert(alert dinkur.Alert) *dinkurapiv1.Alert {
	common := alert.Common()
	a := &dinkurapiv1.Alert{
		Id:      uint64(common.ID),
		Created: convTime(common.CreatedAt),
		Updated: convTime(common.UpdatedAt),
	}
	switch alertType := alert.(type) {
	case dinkur.AlertPlainMessage:
		a.Type = &dinkurapiv1.Alert_PlainMessage{
			PlainMessage: convAlertPlainMessage(alertType),
		}
	case dinkur.AlertAFK:
		a.Type = &dinkurapiv1.Alert_Afk{
			Afk: convAlertAFK(alertType),
		}
	}
	return a
}

func convAlertPlainMessage(alert dinkur.AlertPlainMessage) *dinkurapiv1.AlertPlainMessage {
	return &dinkurapiv1.AlertPlainMessage{
		Message: alert.Message,
	}
}

func convAlertAFK(alert dinkur.AlertAFK) *dinkurapiv1.AlertAfk {
	return &dinkurapiv1.AlertAfk{
		ActiveEntry: convEntryPtr(&alert.ActiveEntry),
		AfkSince:    convTime(alert.AFKSince),
		BackSince:   convTimePtr(alert.BackSince),
	}
}

func convAlertSlice(slice []dinkur.Alert) []*dinkurapiv1.Alert {
	alerts := make([]*dinkurapiv1.Alert, len(slice))
	for i, t := range slice {
		alerts[i] = convAlert(t)
	}
	return alerts
}

func convEvent(ev dinkur.EventType) dinkurapiv1.Event {
	switch ev {
	case dinkur.EventCreated:
		return dinkurapiv1.Event_EVENT_CREATED
	case dinkur.EventUpdated:
		return dinkurapiv1.Event_EVENT_UPDATED
	case dinkur.EventDeleted:
		return dinkurapiv1.Event_EVENT_DELETED
	default:
		return dinkurapiv1.Event_EVENT_UNSPECIFIED
	}
}
