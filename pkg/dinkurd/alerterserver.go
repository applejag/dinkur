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

package dinkurd

import (
	"context"

	dinkurapiv1 "github.com/dinkur/dinkur/api/dinkurapi/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (d *daemon) StreamAlert(*dinkurapiv1.StreamAlertRequest, dinkurapiv1.Alerter_StreamAlertServer) error {
	if err := d.assertConnected(); err != nil {
		return convError(err)
	}
	return status.Error(codes.Unimplemented, "not yet implemented")
}

func (d *daemon) GetAlertList(context.Context, *dinkurapiv1.GetAlertListRequest) (*dinkurapiv1.GetAlertListResponse, error) {
	if err := d.assertConnected(); err != nil {
		return nil, convError(err)
	}
	return nil, status.Error(codes.Unimplemented, "not yet implemented")
}

func (d *daemon) DeleteAlert(context.Context, *dinkurapiv1.DeleteAlertRequest) (*dinkurapiv1.DeleteAlertResponse, error) {
	if err := d.assertConnected(); err != nil {
		return nil, convError(err)
	}
	return nil, status.Error(codes.Unimplemented, "not yet implemented")
}
