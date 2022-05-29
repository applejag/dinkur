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

// Package dinkurclient contains a Dinkur gRPC client implementation.
package dinkurclient

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/dinkur/dinkur/pkg/dinkur"
	"github.com/iver-wharf/wharf-core/v2/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	dinkurapiv1 "github.com/dinkur/dinkur/api/dinkurapi/v1"
)

// Errors that are specific to the Dinkur gRPC client.
var (
	ErrUintTooLarge        = fmt.Errorf("unsigned int value is too large, maximum: %d", uint64(math.MaxUint))
	ErrResponseIsNil       = errors.New("grpc response was nil")
	ErrUnexpectedNilEntry  = errors.New("unexpected nil entry")
	ErrUnexpectedNilStatus = errors.New("unexpected nil status")
)

var log = logger.NewScoped("client")

// Options for the Dinkur client.
type Options struct{}

// NewClient returns a new dinkur.Client-compatible implementation that uses
// gRPC towards a remote Dinkur daemon to perform all dinkur.Client entries.
func NewClient(serverAddr string, opt Options) dinkur.Client {
	return &client{
		Options:    opt,
		serverAddr: serverAddr,
	}
}

type client struct {
	Options
	serverAddr string
	conn       *grpc.ClientConn
	entryer    dinkurapiv1.EntriesClient
	statuses   dinkurapiv1.StatusesClient
}

func (c *client) assertConnected() error {
	if c == nil {
		return dinkur.ErrClientIsNil
	}
	if c.conn == nil || c.entryer == nil || c.statuses == nil {
		return dinkur.ErrNotConnected
	}
	return nil
}

func (c *client) Connect(ctx context.Context) error {
	if c == nil {
		return dinkur.ErrClientIsNil
	}
	if c.conn != nil || c.entryer != nil || c.statuses != nil {
		return dinkur.ErrAlreadyConnected
	}
	// TODO: add credentials via opts args
	conn, err := grpc.DialContext(ctx, c.serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return convError(err)
	}
	c.conn = conn
	c.entryer = dinkurapiv1.NewEntriesClient(conn)
	c.statuses = dinkurapiv1.NewStatusesClient(conn)
	return nil
}

func (c *client) Close() (err error) {
	if conn := c.conn; conn != nil {
		err = conn.Close()
		c.conn = nil
	}
	c.entryer = nil
	return
}

func (c *client) Ping(ctx context.Context) error {
	if err := c.assertConnected(); err != nil {
		return err
	}
	res, err := c.entryer.Ping(ctx, &dinkurapiv1.PingRequest{})
	if err != nil {
		return convError(err)
	}
	if res == nil {
		return ErrResponseIsNil
	}
	return nil
}

func convError(err error) error {
	if err == nil {
		return nil
	}
	s, ok := status.FromError(err)
	if !ok || s == nil {
		return err
	}
	switch s.Code() {
	case codes.NotFound:
		return remessagedErr{s.Message(), dinkur.ErrNotFound}
	default:
		return remessagedErr{fmt.Sprintf("grpc error code %[1]d %[1]q: %[2]s", s.Code(), s.Message()), err}
	}
}

type remessagedErr struct {
	message string
	inner   error
}

func (w remessagedErr) Unwrap() error {
	return w.inner
}

func (w remessagedErr) Is(err error) bool {
	return errors.Is(err, w.inner)
}

func (w remessagedErr) Error() string {
	return w.message
}
