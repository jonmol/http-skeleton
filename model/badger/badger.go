package badger

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/dgraph-io/badger/v4"
	"github.com/jonmol/http-skeleton/model/badger/sillycounter"
	"github.com/jonmol/http-skeleton/util/logging"
)

type Model interface {
	// EnsureDB should be idempotent, making sure the db is in an expected state
	// ie tables/namespaces/indices etc should exist after a call. it should be safe to run at start, so not
	// dropping and recreating tables or such
	EnsureDB(context.Context) error
	// TearDown is destructive and purges all data, useful for integration tests
	TearDown(context.Context) error
	// Close closes anything needed
	Close(context.Context) error
}

type CounterModel interface {
	Model
	IncGlobal(context.Context) (uint64, error)
	IncWord(context.Context, string) (uint64, error)
}

type DB struct {
	db      *badger.DB
	l       *slog.Logger
	Counter CounterModel
	path    string
}

// Close closes the database. The context is there to adher to the shutdown func
// definition
func (db *DB) Close(ctx context.Context) error {
	db.l.Info("Closing the databases")
	if err := db.Counter.Close(ctx); err != nil {
		db.l.Error("Failed to close Counter", logging.Err(err))
	}
	return db.db.Close()
}

// autoclose closes the DB when the context is canceled. It's more of a lifeline
// function since in the case of badger it might take some time and if the shutdown
// is too fast db.Close might not finish before the db is opened again or the server
// shuts down
func (db *DB) autoclose(ctx context.Context) {
	go func() {
		<-ctx.Done()
		if db.db != nil && !db.db.IsClosed() {
			db.l.Info("Closing the databases (autoclose)")
			if err := db.Close(ctx); err != nil {
				db.l.Error("Failed to properly close the database", logging.Err(err))
			}
		}
	}()
}

func (db *DB) EnsureDB(ctx context.Context) error {
	return db.Counter.EnsureDB(ctx)
}

func (db *DB) Healthy(_ context.Context) bool {
	return !db.db.IsClosed()
}

// TearDown doesn't need to cascade on a DB level since it can just Drop all and delete the
// DB file
func (db *DB) TearDown(_ context.Context) error {
	err := db.db.DropAll()
	err = errors.Join(err, db.db.Close())
	err = errors.Join(err, os.RemoveAll(db.path))
	return err
}

func (db *DB) Open(_ context.Context) error {
	dbOpts := badger.DefaultOptions(db.path).WithLoggingLevel(badger.INFO).WithLogger(NewLogger())
	d, err := badger.Open(dbOpts)
	if err != nil {
		return err
	}
	db.db = d
	db.Counter = sillycounter.New(d)
	return nil
}

func New(ctx context.Context, p string) *DB {
	l := slog.With(logging.Lib("model"))
	m := DB{
		l:    l,
		path: p,
	}
	m.autoclose(ctx)
	return &m
}
