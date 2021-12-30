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
	"errors"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/dinkur/dinkur/pkg/dinkur"
	"github.com/mattn/go-colorable"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func nilNotFoundError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	return err
}

type Options struct {
	MkdirAll             bool
	SkipMigrateOnConnect bool
	DebugLogging         bool
	DebugColorful        bool
}

func NewClient(dsn string, opt Options) dinkur.Client {
	return &client{Options: opt, sqliteDsn: dsn}
}

type client struct {
	Options
	sqliteDsn     string
	db            *gorm.DB
	prevMigStatus dinkur.MigrationStatus
}

func (c *client) assertConnected() error {
	if c == nil {
		return dinkur.ErrClientIsNil
	}
	if c.db == nil {
		return dinkur.ErrNotConnected
	}
	return nil
}

func (c *client) Connect() error {
	if c == nil {
		return dinkur.ErrClientIsNil
	}
	if c.db != nil {
		return dinkur.ErrAlreadyConnected
	}
	if c.MkdirAll {
		dir := filepath.Dir(c.sqliteDsn)
		os.MkdirAll(dir, os.ModeDir)
	}
	var err error
	c.db, err = gorm.Open(sqlite.Open(c.sqliteDsn), &gorm.Config{
		Logger: getLogger(c.Options),
	})
	if err != nil {
		return err
	}
	sqlDB, err := c.db.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxOpenConns(1)
	if !c.SkipMigrateOnConnect {
		return c.Migrate()
	}
	return nil
}

func getLogger(opt Options) logger.Interface {
	if opt.DebugLogging {
		return logger.New(log.New(colorable.NewColorableStderr(), "\r\n", log.LstdFlags), logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: false,
			Colorful:                  opt.DebugColorful,
		})
	}
	return logger.Discard
}

func (c *client) Ping() error {
	if err := c.assertConnected(); err != nil {
		return err
	}
	sql, err := c.db.DB()
	if err != nil {
		return err
	}
	return sql.Ping()
}

func (c *client) Close() error {
	if err := c.assertConnected(); err != nil {
		return err
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
