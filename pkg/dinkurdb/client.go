// Dinkur the task time tracking utility.
// <https://github.com/dinkur/dinkur>
//
// Copyright (C) 2021 Kalle Fagerberg
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

package dinkurdb

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Client interface {
	Connect(filename string) error
	Ping() error
	Close() error

	Migrate() error
	MigrationStatus() (MigrationStatus, error)

	GetTask(id uint) (Task, error)
	StartTask(task NewTask) (StartedTask, error)
	ActiveTask() (*Task, error)
	StopActiveTask() (*Task, error)
}

func NewClient() Client {
	return &client{}
}

type client struct {
	db            *gorm.DB
	prevMigStatus MigrationStatus
}

func (c *client) Connect(sqliteDSN string) error {
	if c.db != nil {
		return ErrAlreadyConnected
	}
	var err error
	c.db, err = gorm.Open(sqlite.Open(sqliteDSN), &gorm.Config{
		Logger: logger.Discard,
		//Logger: logger.New(log.New(colorable.NewColorableStdout(), "\r\n", log.LstdFlags), logger.Config{
		//	SlowThreshold:             200 * time.Millisecond,
		//	LogLevel:                  logger.Info,
		//	IgnoreRecordNotFoundError: false,
		//	Colorful:                  true,
		//}),
	})
	if err != nil {
		return err
	}
	sqlDB, err := c.db.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxOpenConns(1)
	return nil
}

func (c *client) Ping() error {
	if c.db == nil {
		return ErrNotConnected
	}
	sql, err := c.db.DB()
	if err != nil {
		return err
	}
	return sql.Ping()
}

func (c *client) Close() error {
	if c.db == nil {
		return ErrNotConnected
	}
	sql, err := c.db.DB()
	if err != nil {
		return err
	}
	if err := sql.Close(); err != nil {
		return err
	}
	c.db = nil
	return nil
}

func (c *client) transaction(f func(tx *client) error) error {
	return c.db.Transaction(func(tx *gorm.DB) error {
		newClient := *c
		newClient.db = tx
		return f(&newClient)
	})
}
