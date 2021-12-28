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
