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

package dinkurd

import (
	"context"
	"time"

	dinkurapiv1 "github.com/dinkur/dinkur/api/dinkurapi/v1"
	"github.com/dinkur/dinkur/pkg/dinkur"
	"github.com/dinkur/dinkur/pkg/fromgrpc"
	"github.com/dinkur/dinkur/pkg/togrpc"
	"gopkg.in/typ.v2"
)

func (d *daemon) StreamStatus(req *dinkurapiv1.StreamStatusRequest, stream dinkurapiv1.Statuses_StreamStatusServer) error {
	if err := d.assertConnected(); err != nil {
		return convError(err)
	}
	if req == nil {
		return convError(ErrRequestIsNil)
	}
	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()
	ch, err := d.client.StreamStatus(ctx)
	if err != nil {
		return convError(err)
	}
	for ev := range ch {
		if err := stream.Send(&dinkurapiv1.StreamStatusResponse{
			Status: togrpc.Status(ev.Status),
		}); err != nil {
			return convError(err)
		}
	}
	return nil
}

func (d *daemon) SetStatus(ctx context.Context, req *dinkurapiv1.SetStatusRequest) (*dinkurapiv1.SetStatusResponse, error) {
	if err := d.assertConnected(); err != nil {
		return nil, convError(err)
	}
	if req == nil {
		return nil, convError(ErrRequestIsNil)
	}
	err := d.client.SetStatus(ctx, dinkur.EditStatus{
		AFKSince:  fromgrpc.TimePtr(req.AfkSince),
		BackSince: fromgrpc.TimePtr(req.BackSince),
	})
	if err != nil {
		return nil, convError(err)
	}
	return &dinkurapiv1.SetStatusResponse{}, nil
}

func (d *daemon) GetStatus(ctx context.Context, req *dinkurapiv1.GetStatusRequest) (*dinkurapiv1.GetStatusResponse, error) {
	if err := d.assertConnected(); err != nil {
		return nil, convError(err)
	}
	if req == nil {
		return nil, convError(ErrRequestIsNil)
	}
	status, err := d.client.GetStatus(ctx)
	if err != nil {
		return nil, convError(err)
	}
	return &dinkurapiv1.GetStatusResponse{
		Status: togrpc.Status(status),
	}, nil
}

func (d *daemon) markAsNotAFK(ctx context.Context) {
	lastStatus := d.lastStatus
	if lastStatus.AFKSince == nil && lastStatus.BackSince == nil {
		return
	}
	newStatus := dinkur.EditStatus{
		AFKSince:  nil,
		BackSince: nil,
	}
	d.client.SetStatus(ctx, newStatus)
	d.lastStatus = newStatus
}

func (d *daemon) markAsReturnedFromAFK(ctx context.Context) {
	lastStatus := d.lastStatus
	if lastStatus.AFKSince != nil && lastStatus.BackSince != nil {
		return
	}
	newStatus := dinkur.EditStatus{
		AFKSince:  lastStatus.AFKSince,
		BackSince: typ.Ref(time.Now()),
	}
	if newStatus.AFKSince == nil {
		newStatus.AFKSince = typ.Ref(time.Now())
	}
	d.client.SetStatus(ctx, newStatus)
	d.lastStatus = newStatus
}

func (d *daemon) markAsAFK(ctx context.Context) {
	lastStatus := d.lastStatus
	if lastStatus.AFKSince != nil && lastStatus.BackSince == nil {
		return
	}
	newStatus := dinkur.EditStatus{
		AFKSince:  lastStatus.AFKSince,
		BackSince: nil,
	}
	if newStatus.AFKSince == nil {
		newStatus.AFKSince = typ.Ref(time.Now())
	}
	d.client.SetStatus(ctx, newStatus)
	d.lastStatus = newStatus
}
