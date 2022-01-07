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
	"time"

	"github.com/dinkur/dinkur/pkg/dinkur"
	"github.com/dinkur/dinkur/pkg/timeutil"
	"github.com/iver-wharf/wharf-core/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	dinkurapiv1 "github.com/dinkur/dinkur/api/dinkurapi/v1"
)

// Errors that are specific to the Dinkur gRPC client.
var (
	ErrUintTooLarge       = fmt.Errorf("unsigned int value is too large, maximum: %d", uint64(math.MaxUint))
	ErrResponseIsNil      = errors.New("grpc response was nil")
	ErrUnexpectedNilTask  = errors.New("unexpected nil task")
	ErrUnexpectedNilAlert = errors.New("unexpected nil alert")
)

var log = logger.NewScoped("client")

// Options for the Dinkur client.
type Options struct{}

// NewClient returns a new dinkur.Client-compatible implementation that uses
// gRPC towards a remote Dinkur daemon to perform all dinkur.Client tasks.
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
	tasker     dinkurapiv1.TaskerClient
	alerter    dinkurapiv1.AlerterClient
}

func (c *client) assertConnected() error {
	if c == nil {
		return dinkur.ErrClientIsNil
	}
	if c.conn == nil || c.tasker == nil || c.alerter == nil {
		return dinkur.ErrNotConnected
	}
	return nil
}

func (c *client) Connect(ctx context.Context) error {
	if c == nil {
		return dinkur.ErrClientIsNil
	}
	if c.conn != nil || c.tasker != nil || c.alerter != nil {
		return dinkur.ErrAlreadyConnected
	}
	// TODO: add credentials via opts args
	conn, err := grpc.DialContext(ctx, c.serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return convError(err)
	}
	c.conn = conn
	c.tasker = dinkurapiv1.NewTaskerClient(conn)
	c.alerter = dinkurapiv1.NewAlerterClient(conn)
	return nil
}

func (c *client) Close() (err error) {
	if conn := c.conn; conn != nil {
		err = conn.Close()
		c.conn = nil
	}
	c.tasker = nil
	return
}

func (c *client) Ping(ctx context.Context) error {
	if err := c.assertConnected(); err != nil {
		return err
	}
	res, err := c.tasker.Ping(ctx, &dinkurapiv1.PingRequest{})
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

func uint64ToUint(v uint64) (uint, error) {
	if v > math.MaxUint {
		return 0, ErrUintTooLarge
	}
	return uint(v), nil
}

func convStringPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
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
	t := ts.AsTime().Local()
	return &t
}

func convTimestampOrZero(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}
	return ts.AsTime().Local()
}

func convShorthand(s timeutil.TimeSpanShorthand) dinkurapiv1.GetTaskListRequest_Shorthand {
	switch s {
	case timeutil.TimeSpanPast:
		return dinkurapiv1.GetTaskListRequest_PAST
	case timeutil.TimeSpanFuture:
		return dinkurapiv1.GetTaskListRequest_FUTURE
	case timeutil.TimeSpanThisDay:
		return dinkurapiv1.GetTaskListRequest_THIS_DAY
	case timeutil.TimeSpanThisWeek:
		return dinkurapiv1.GetTaskListRequest_THIS_MON_TO_SUN
	case timeutil.TimeSpanPrevDay:
		return dinkurapiv1.GetTaskListRequest_PREV_DAY
	case timeutil.TimeSpanPrevWeek:
		return dinkurapiv1.GetTaskListRequest_PREV_MON_TO_SUN
	case timeutil.TimeSpanNextDay:
		return dinkurapiv1.GetTaskListRequest_NEXT_DAY
	case timeutil.TimeSpanNextWeek:
		return dinkurapiv1.GetTaskListRequest_NEXT_MON_TO_SUN
	default:
		return dinkurapiv1.GetTaskListRequest_NONE
	}
}

func convTaskPtr(task *dinkurapiv1.Task) (*dinkur.Task, error) {
	if task == nil {
		return nil, nil
	}
	id, err := uint64ToUint(task.Id)
	if err != nil {
		return nil, fmt.Errorf("convert task ID: %w", err)
	}
	return &dinkur.Task{
		CommonFields: dinkur.CommonFields{
			ID:        id,
			CreatedAt: convTimestampOrZero(task.CreatedAt),
			UpdatedAt: convTimestampOrZero(task.UpdatedAt),
		},
		Name:  task.Name,
		Start: convTimestampOrZero(task.Start),
		End:   convTimestampPtr(task.End),
	}, nil
}

func convTaskPtrNoNil(task *dinkurapiv1.Task) (dinkur.Task, error) {
	t, err := convTaskPtr(task)
	if err != nil {
		return dinkur.Task{}, err
	}
	if t == nil {
		return dinkur.Task{}, ErrUnexpectedNilTask
	}
	return *t, nil
}

func convTaskSlice(slice []*dinkurapiv1.Task) ([]dinkur.Task, error) {
	tasks := make([]dinkur.Task, 0, len(slice))
	for _, t := range slice {
		t2, err := convTaskPtr(t)
		if t2 == nil {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("task #%d %q: %w", t.Id, t.Name, err)
		}
		tasks = append(tasks, *t2)
	}
	return tasks, nil
}

func convAlertPtr(alert *dinkurapiv1.Alert) (*dinkur.Alert, error) {
	if alert == nil {
		return nil, nil
	}
	id, err := uint64ToUint(alert.Id)
	if err != nil {
		return nil, err
	}
	a := dinkur.Alert{
		CommonFields: dinkur.CommonFields{
			ID:        id,
			CreatedAt: convTimestampOrZero(alert.CreatedAt),
			UpdatedAt: convTimestampOrZero(alert.UpdatedAt),
		},
	}
	switch alertType := alert.Type.(type) {
	case *dinkurapiv1.Alert_PlainMessage:
		a.Type = convAlertPlainMessage(alertType.PlainMessage)
	case *dinkurapiv1.Alert_Afk:
		at, err := convAlertAFK(alertType.Afk)
		if err != nil {
			return nil, err
		}
		a.Type = at
	case *dinkurapiv1.Alert_FormerlyAfk:
		at, err := convAlertFormerlyAFK(alertType.FormerlyAfk)
		if err != nil {
			return nil, err
		}
		a.Type = at
	}
	return &a, nil
}

func convAlertPlainMessage(alert *dinkurapiv1.AlertPlainMessage) dinkur.AlertType {
	if alert == nil {
		return nil
	}
	return dinkur.AlertPlainMessage{
		Message: alert.Message,
	}
}

func convAlertAFK(alert *dinkurapiv1.AlertAfk) (dinkur.AlertType, error) {
	if alert == nil {
		return nil, nil
	}
	task, err := convTaskPtrNoNil(alert.ActiveTask)
	if err != nil {
		return nil, err
	}
	return dinkur.AlertAFK{
		ActiveTask: task,
	}, nil
}

func convAlertFormerlyAFK(alert *dinkurapiv1.AlertFormerlyAfk) (dinkur.AlertType, error) {
	if alert == nil {
		return nil, nil
	}
	task, err := convTaskPtrNoNil(alert.ActiveTask)
	if err != nil {
		return nil, err
	}
	return dinkur.AlertFormerlyAFK{
		ActiveTask: task,
		AFKSince:   convTimestampOrZero(alert.AfkSince),
	}, nil
}

func convAlertSlice(slice []*dinkurapiv1.Alert) ([]dinkur.Alert, error) {
	tasks := make([]dinkur.Alert, 0, len(slice))
	for _, a := range slice {
		a2, err := convAlertPtr(a)
		if a2 == nil {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("alert #%d: %w", a.Id, err)
		}
		tasks = append(tasks, *a2)
	}
	return tasks, nil
}

func convEvent(ev dinkurapiv1.Event) dinkur.EventType {
	switch ev {
	case dinkurapiv1.Event_CREATED:
		return dinkur.EventCreated
	case dinkurapiv1.Event_UPDATED:
		return dinkur.EventUpdated
	case dinkurapiv1.Event_DELETED:
		return dinkur.EventDeleted
	default:
		return dinkur.EventUnknown
	}
}
