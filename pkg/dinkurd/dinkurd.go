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
	"errors"
	"fmt"
	"math"
	"net"
	"time"

	dinkurapiv1 "github.com/dinkur/dinkur/api/dinkurapi/v1"
	"github.com/dinkur/dinkur/pkg/dinkur"
	"github.com/dinkur/dinkur/pkg/timeutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	ErrUintTooLarge      = fmt.Errorf("unsigned int value is too large, maximum: %d", uint64(math.MaxUint))
	ErrDaemonIsNil       = errors.New("daemon is nil")
	ErrTaskerServerIsNil = errors.New("tasker server is nil")
	ErrAlreadyServing    = errors.New("daemon instance is already running")
)

func convError(err error) error {
	switch {
	case errors.Is(err, dinkur.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, ErrUintTooLarge),
		errors.Is(err, dinkur.ErrLimitTooLarge):
		return status.Error(codes.OutOfRange, err.Error())
	case errors.Is(err, dinkur.ErrTaskEndBeforeStart),
		errors.Is(err, dinkur.ErrTaskNameEmpty):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, ErrTaskerServerIsNil),
		errors.Is(err, dinkur.ErrNotConnected),
		errors.Is(err, dinkur.ErrAlreadyConnected),
		errors.Is(err, dinkur.ErrClientIsNil):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return err
	}
}

func uint64ToUint(v uint64) (uint, error) {
	if v > math.MaxUint {
		return 0, ErrUintTooLarge
	}
	return uint(v), nil
}

type Options struct {
	Host string
	Port uint16
}

var DefaultOptions = Options{
	Host: "localhost",
	Port: 59122,
}

type Daemon interface {
	Serve(ctx context.Context) error
	Close() error
}

func NewDaemon(client dinkur.Client, opt Options) Daemon {
	if opt.Host == "" {
		opt.Host = DefaultOptions.Host
	}
	if opt.Port == 0 {
		opt.Port = DefaultOptions.Port
	}
	return &daemon{
		Options: opt,
		tasker:  NewTaskerServer(client),
	}
}

type daemon struct {
	Options
	tasker dinkurapiv1.TaskerServer

	grpcServer *grpc.Server
	listener   net.Listener
}

func (d *daemon) Serve(ctx context.Context) error {
	if d == nil {
		return ErrDaemonIsNil
	}
	if d.tasker == nil {
		return ErrTaskerServerIsNil
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
	dinkurapiv1.RegisterTaskerServer(grpcServer, d.tasker)
	return grpcServer.Serve(lis)
}

func (d *daemon) Close() (err error) {
	if srv := d.grpcServer; srv != nil {
		srv.GracefulStop()
	}
	if lis := d.listener; lis != nil {
		err = lis.Close()
	}
	d.grpcServer = nil
	d.listener = nil
	return
}

func convTaskPtr(task *dinkur.Task) *dinkurapiv1.Task {
	if task == nil {
		return nil
	}
	return &dinkurapiv1.Task{
		Id:        uint64(task.ID),
		CreatedAt: convTime(task.CreatedAt),
		UpdatedAt: convTime(task.UpdatedAt),
		Name:      task.Name,
		Start:     convTime(task.Start),
		End:       convTimePtr(task.End),
	}
}

func convTaskSlice(slice []dinkur.Task) []*dinkurapiv1.Task {
	tasks := make([]*dinkurapiv1.Task, len(slice))
	for i, t := range slice {
		tasks[i] = convTaskPtr(&t)
	}
	return tasks
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

func convShorthand(s dinkurapiv1.GetTaskListRequest_Shorthand) timeutil.TimeSpanShorthand {
	switch s {
	case dinkurapiv1.GetTaskListRequest_PAST:
		return timeutil.TimeSpanPast
	case dinkurapiv1.GetTaskListRequest_FUTURE:
		return timeutil.TimeSpanFuture
	case dinkurapiv1.GetTaskListRequest_THIS_DAY:
		return timeutil.TimeSpanThisDay
	case dinkurapiv1.GetTaskListRequest_THIS_MON_TO_SUN:
		return timeutil.TimeSpanThisWeek
	case dinkurapiv1.GetTaskListRequest_PREV_DAY:
		return timeutil.TimeSpanPrevDay
	case dinkurapiv1.GetTaskListRequest_PREV_MON_TO_SUN:
		return timeutil.TimeSpanPrevWeek
	case dinkurapiv1.GetTaskListRequest_NEXT_DAY:
		return timeutil.TimeSpanNextDay
	case dinkurapiv1.GetTaskListRequest_NEXT_MON_TO_SUN:
		return timeutil.TimeSpanNextWeek
	default:
		return timeutil.TimeSpanNone
	}
}
