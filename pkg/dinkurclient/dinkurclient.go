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

// Package dinkurclient contains a Dinkur gRPC client implementation.
package dinkurclient

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/dinkur/dinkur/pkg/dinkur"
	"github.com/dinkur/dinkur/pkg/timeutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	dinkurapiv1 "github.com/dinkur/dinkur/api/dinkurapi/v1"
)

// Errors that are specific to the Dinkur gRPC client.
var (
	ErrUintTooLarge      = fmt.Errorf("unsigned int value is too large, maximum: %d", uint64(math.MaxUint))
	ErrResponseIsNil     = errors.New("grpc response was nil")
	ErrUnexpectedNilTask = errors.New("unexpected nil task")
)

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

func convUintPtr(i *uint) uint64 {
	if i == nil {
		return 0
	}
	return uint64(*i)
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
