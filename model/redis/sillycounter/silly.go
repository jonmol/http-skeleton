package sillycounter

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jonmol/http-skeleton/model/redis/common"
	"github.com/jonmol/http-skeleton/util/logging"
	"github.com/redis/go-redis/v9"
)

const (
	prefix        = "a"
	globalCounter = prefix + "globalC"
)

type SillyCounter struct {
	db *redis.Client
	l  *slog.Logger
}

func New(db *redis.Client) *SillyCounter {
	return &SillyCounter{
		db: db,
		l:  slog.With(logging.Lib("badger.sillycounter")),
	}
}

// TearDown deletes all keys in the db with our prefix
func (s *SillyCounter) TearDown(ctx context.Context) error {
	i, err := common.DeleteAll(ctx, s.db, prefix)
	s.l.Debug("Deleted in teardown", slog.Int64("deleted", i))
	return err
}

// nothing to do here
func (s *SillyCounter) EnsureDB(_ context.Context) error {
	return nil
}

// nothing to do here
func (s *SillyCounter) Close(_ context.Context) error {
	return nil
}

func (s *SillyCounter) IncGlobal(ctx context.Context) (uint64, error) {
	return s.incr(ctx, globalCounter)
}

func (s *SillyCounter) IncWord(ctx context.Context, w string) (uint64, error) {
	return s.incr(ctx, fmt.Sprintf("%s%s", prefix, w))
}

func (s *SillyCounter) incr(ctx context.Context, k string) (uint64, error) {
	res, err := s.db.Incr(ctx, k).Result()
	if res > 0 {
		return uint64(res), err
	}
	return 0, err
}
