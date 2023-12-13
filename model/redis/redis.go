package redis

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jonmol/http-skeleton/model/redis/sillycounter"
	"github.com/jonmol/http-skeleton/util/logging"
	"github.com/redis/go-redis/v9"
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
type DB struct {
	db      *redis.Client
	l       *slog.Logger
	addr    string
	pass    string
	Counter *sillycounter.SillyCounter
}

func (db *DB) Close(_ context.Context) error {
	return db.db.Close()
}

func (db *DB) EnsureDB(_ context.Context) error {
	return nil
}

func (db *DB) TearDown(_ context.Context) error {
	return errors.New("not implemented")
}

func (db *DB) Healthy(ctx context.Context) bool {
	if err := db.db.Ping(ctx).Err(); err != nil {
		return false
	}
	return true
}

func (db *DB) Open(ctx context.Context) error {
	client := redis.NewClient(&redis.Options{
		Addr:     db.addr,
		Password: db.pass,
	})

	if st := client.Ping(ctx); st.Err() != nil {
		return st.Err()
	}

	db.Counter = sillycounter.New(client)
	db.db = client
	return nil
}

func (db *DB) autoclose(ctx context.Context) {
	go func() {
		<-ctx.Done()
		db.l.Info("Closing the databases (autoclose)")
		if err := db.Close(ctx); err != nil {
			db.l.Error("Failed to properly close the database", logging.Err(err))
		}
	}()
}

func New(ctx context.Context, addr, pass string) *DB {
	l := slog.With(logging.Lib("model"))
	db := DB{
		l:    l,
		addr: addr,
		pass: pass,
	}
	db.autoclose(ctx)
	return &db
}
