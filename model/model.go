package model

import (
	"context"
	"log/slog"

	"github.com/jonmol/http-skeleton/model/badger"
	"github.com/jonmol/http-skeleton/model/redis"
	"github.com/jonmol/http-skeleton/util/logging"
)

// db is a high level representation of some database. It contains a few functions that makes sense, but more
// should be added if your DB is more advanced
type db interface {
	// Open connects to the database
	Open(context.Context) error

	// EnsureDB should be idempotent, making sure the db is in an expected state
	// ie tables/namespaces/indices etc should exist after a call. it should be safe to run at start, so not
	// dropping and recreating tables or such
	EnsureDB(context.Context) error

	// TearDown is destructive and purges all data, useful for integration tests
	TearDown(context.Context) error

	// Close closes anything needed to be able to have a clean shutdown
	Close(ctx context.Context) error

	// Healthy checks if the DB is healthy
	Healthy(ctx context.Context) bool
}

// Counter represents a Counter, it could be using MariDB, badger, bolt, redis or any kind of database.
// it will be transparent to the consumber
type Counter interface {
	IncGlobal(context.Context) (uint64, error)
	IncWord(context.Context, string) (uint64, error)
}

type DB struct {
	db      db
	l       *slog.Logger
	Counter Counter
}

func (db *DB) Close(ctx context.Context) error {
	return db.db.Close(ctx)
}

func (db *DB) EnsureDB(ctx context.Context) error {
	return db.db.EnsureDB(ctx)
}

func (db *DB) TearDown(ctx context.Context) error {
	return db.db.TearDown(ctx)
}

func (db *DB) Healthy(ctx context.Context) bool {
	return db.db.Healthy(ctx)
}

func (db *DB) OpenBadger(ctx context.Context, p string) error {
	badgerC := badger.New(ctx, p)
	if err := badgerC.Open(ctx); err != nil {
		return err
	}
	db.db = badgerC
	db.Counter = badgerC.Counter

	return nil
}

func (db *DB) OpenRedis(ctx context.Context, addr, pass string) error {
	redisC := redis.New(ctx, addr, pass)
	if err := redisC.Open(ctx); err != nil {
		return err
	}
	db.db = redisC
	db.Counter = redisC.Counter

	return nil
}

func NewModel(_ context.Context) *DB {
	l := slog.With(logging.Lib("model"))
	m := DB{
		l: l,
	}
	return &m
}
