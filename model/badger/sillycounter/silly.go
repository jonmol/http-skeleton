package sillycounter

import (
	"context"
	"errors"
	"log/slog"

	"github.com/dgraph-io/badger/v4"
	"github.com/jonmol/http-skeleton/util/logging"
)

var (
	prefix        = []byte("a")
	globalCounter = append(prefix, []byte("globC")...)
)

type SillyCounter struct {
	db *badger.DB
	gc *badger.Sequence
	l  *slog.Logger
}

func New(db *badger.DB) *SillyCounter {
	return &SillyCounter{
		db: db,
		l:  slog.With(logging.Lib("badger.sillycounter")),
	}
}

func (s *SillyCounter) EnsureDB(_ context.Context) error {
	seq, err := s.db.GetSequence(globalCounter, 100)
	if err != nil {
		return err
	}
	s.gc = seq
	s.l.Debug("Added counter")
	return nil
}

// TearDown deletes all keys in the db with our prefix
func (s *SillyCounter) TearDown(_ context.Context) error {
	return s.db.DropPrefix(prefix)
}

func (s *SillyCounter) Close(_ context.Context) error {
	if s.gc != nil {
		return s.gc.Release()
	}
	return nil
}

func (s *SillyCounter) IncGlobal(_ context.Context) (uint64, error) {
	if s.gc == nil {
		return 0, errors.New("global counter nil")
	}
	return humanize(s.gc.Next())
}

func (s *SillyCounter) IncWord(_ context.Context, w string) (uint64, error) {
	seq, err := s.db.GetSequence(append(prefix, []byte(w)...), 2)
	if err != nil {
		return 0, err
	}
	defer func() {
		err = errors.Join(err, seq.Release())
	}()
	res, err := humanize(seq.Next())
	return res, err
}

func humanize(i uint64, e error) (uint64, error) {
	return i + 1, e
}
