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

package dinkurclient

import (
	"context"
	"io"

	dinkurapiv1 "github.com/dinkur/dinkur/api/dinkurapi/v1"
	v1 "github.com/dinkur/dinkur/api/dinkurapi/v1"
	"github.com/dinkur/dinkur/pkg/dinkur"
	"github.com/dinkur/dinkur/pkg/fromgrpc"
	"github.com/dinkur/dinkur/pkg/togrpc"
)

func (c *client) StreamStatus(ctx context.Context) (<-chan dinkur.StreamedStatus, error) {
	if err := c.assertConnected(); err != nil {
		return nil, err
	}
	stream, err := c.statuses.StreamStatus(ctx, &dinkurapiv1.StreamStatusRequest{})
	if err != nil {
		return nil, convError(err)
	}
	statusChan := make(chan dinkur.StreamedStatus)
	go func() {
		for {
			res, err := stream.Recv()
			if err != nil {
				if err != io.EOF {
					log.Error().
						WithError(convError(err)).
						Message("Error when streaming statuses. Closing stream.")
				}
				close(statusChan)
				return
			}
			if res == nil {
				continue
			}
			const logWarnMsg = "Error when streaming statuses. Ignoring message."
			status, err := fromgrpc.StatusPtrNoNil(res.Status)
			if err != nil {
				log.Warn().WithError(convError(err)).
					Message(logWarnMsg)
				continue
			}
			statusChan <- dinkur.StreamedStatus{
				Status: status,
			}
		}
	}()
	return statusChan, nil
}

func (c *client) SetStatus(ctx context.Context, edit dinkur.EditStatus) error {
	_, err := invoke(ctx, c, c.statuses.SetStatus, &v1.SetStatusRequest{
		AfkSince:  togrpc.TimestampPtr(edit.AFKSince),
		BackSince: togrpc.TimestampPtr(edit.BackSince),
	})
	return err
}

func (c *client) GetStatus(ctx context.Context) (dinkur.Status, error) {
	res, err := invoke(ctx, c, c.statuses.GetStatus, &v1.GetStatusRequest{})
	if err != nil {
		return dinkur.Status{}, err
	}
	return fromgrpc.StatusPtrNoNil(res.Status)
}
