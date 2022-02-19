// Dinkur the task time tracking utility.
// <https://github.com/dinkur/dinkur>
//
// Copyright (C) 2021 Kalle Fagerberg
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

// Package dinkurdb contains a dinkur.Client implementation that targets an
// Sqlite3 database file.
package dinkurdb

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/dinkur/dinkur/pkg/dbmodel"
	"github.com/dinkur/dinkur/pkg/dinkur"
	"github.com/iver-wharf/wharf-core/pkg/gormutil"
	"github.com/iver-wharf/wharf-core/pkg/logger"
	"gopkg.in/typ.v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var log = logger.NewScoped("DB")

func nilNotFoundError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	return err
}

// Options for the client's Sqlite3 database connection.
type Options struct {
	// MkdirAll lets the client create any containing directory for where the
	// database file is to be stored upon connecting to it. If set to false and
	// the containing directory does not exist, then the Connect method will
	// return a "file not found" error.
	MkdirAll bool
	// SkipMigrateOnConnect disables the migration check done when the Connect
	// method is invoked.
	SkipMigrateOnConnect bool
	// DebugLogging enables logging of SQL queries and warnings issued by
	// GORM.
	DebugLogging bool
}

// NewClient creates a new dinkur.Client-compatible client that uses an Sqlite3
// database file for persistence.
func NewClient(dsn string, opt Options) dinkur.Client {
	return &client{
		Options:   opt,
		sqliteDsn: dsn,
		entryObs: &typ.Publisher[entryEvent]{
			PubTimeoutAfter: 10 * time.Second,
			OnPubTimeout: func(ev entryEvent) {
				log.Warn().
					WithUint("id", ev.dbEntry.ID).
					WithString("name", ev.dbEntry.Name).
					WithStringer("event", ev.event).
					Message("Timed out sending entry event.")
			},
		},
		statusObs: &typ.Publisher[statusEvent]{
			PubTimeoutAfter: 10 * time.Second,
			OnPubTimeout: func(ev statusEvent) {
				log.Warn().Message("Timed out sending status event.")
			},
		},
	}
}

type client struct {
	Options
	sqliteDsn      string
	db             *gorm.DB
	prevMigChecked bool
	prevMigVersion dbmodel.MigrationVersion
	entryObs       *typ.Publisher[entryEvent]
	statusObs      *typ.Publisher[statusEvent]
}

type entryEvent struct {
	dbEntry dbmodel.Entry
	event   dinkur.EventType
}

type statusEvent struct {
	dbStatus dbmodel.Status
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

func (c *client) Connect(ctx context.Context) error {
	if c == nil {
		return dinkur.ErrClientIsNil
	}
	if c.db != nil {
		return dinkur.ErrAlreadyConnected
	}
	if c.MkdirAll {
		dir := filepath.Dir(c.sqliteDsn)
		os.MkdirAll(dir, 0700)
	}
	var err error
	c.db, err = gorm.Open(sqlite.Open(c.sqliteDsn), &gorm.Config{
		Logger: getLogger(c.Options),
	})
	if err != nil {
		return err
	}
	if err := c.db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		return err
	}
	sqlDB, err := c.db.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxOpenConns(1)
	if !c.SkipMigrateOnConnect {
		return c.Migrate(ctx)
	}
	return nil
}

func getLogger(opt Options) gormlogger.Interface {
	if opt.DebugLogging {
		return gormutil.DefaultLogger
	}
	return gormlogger.Discard
}

func (c *client) Ping(ctx context.Context) error {
	if err := c.assertConnected(); err != nil {
		return err
	}
	sql, err := c.db.DB()
	if err != nil {
		return err
	}
	return sql.PingContext(ctx)
}

func (c *client) Close() error {
	if err := c.assertConnected(); err != nil {
		return err
	}
	if err := c.entryObs.UnsubAll(); err != nil {
		log.Warn().WithError(err).Message("Failed to unsub all entry subscriptions.")
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

func (c *client) withContext(ctx context.Context) *client {
	newClient := *c
	newClient.db = newClient.db.WithContext(ctx)
	return &newClient
}
