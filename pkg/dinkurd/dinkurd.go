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
	"github.com/iver-wharf/wharf-core/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/typ.v2"
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
	oldAFKAlert *dinkur.AlertAFK
}

func (d *daemon) onEntryMutation(ctx context.Context) {
	d.oldAFKAlert = nil
	go d.client.DeleteAlertByType(ctx, dinkur.AlertTypeAFK)
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
	d.updateAFKStatusAsWeAreStarting(ctx)
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
	d.updateAFKStatusAsWeAreClosing()
	return
}

func (d *daemon) updateAFKStatusAsWeAreStarting(ctx context.Context) {
	alerts, err := d.client.GetAlertList(ctx)
	if err != nil {
		return
	}
	afk, ok := findAFKAlert(alerts)
	if !ok {
		return
	}
	afk.BackSince = typ.Ref(time.Now())
	d.client.UpdateAlert(context.Background(), dinkur.EditAlert{
		ID:    afk.ID,
		Alert: afk,
	})
}

func findAFKAlert(alerts []dinkur.Alert) (dinkur.AlertAFK, bool) {
	for _, a := range alerts {
		afk, ok := a.(dinkur.AlertAFK)
		if ok {
			return afk, true
		}
	}
	return dinkur.AlertAFK{}, false
}

func (d *daemon) updateAFKStatusAsWeAreClosing() {
	// must use new context as base context from Serve is cancelled by now
	entry, err := d.client.GetActiveEntry(context.Background())
	if err != nil || entry == nil {
		return
	}
	d.client.CreateOrUpdateAlertByType(context.Background(), dinkur.NewAlert{
		Alert: dinkur.AlertAFK{
			ActiveEntry: *entry,
			AFKSince:    time.Now(),
		},
	})
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
			updated, err := d.client.CreateOrUpdateAlertByType(ctx, dinkur.NewAlert{
				Alert: dinkur.AlertAFK{
					AFKSince:    time.Now(),
					ActiveEntry: *entry,
				},
			})
			if err != nil {
				log.Warn().WithError(err).
					Message("Failed to create AFK alert.")
				continue
			}
			d.oldAFKAlert = typ.Ref(updated.After.(dinkur.AlertAFK))
		case <-stoppedChan:
			var alert dinkur.AlertAFK
			if oldAFKAlert := d.oldAFKAlert; oldAFKAlert != nil {
				alert = *oldAFKAlert
			}
			alert.BackSince = typ.Ref(time.Now())
			updated, err := d.client.CreateOrUpdateAlertByType(ctx, dinkur.NewAlert{
				Alert: alert,
			})
			if err != nil {
				log.Warn().WithError(err).
					Message("Failed to set as formerly AFK.")
				continue
			}
			d.oldAFKAlert = typ.Ref(updated.After.(dinkur.AlertAFK))
		case <-done:
			return
		}
	}
}
