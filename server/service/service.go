package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jonmol/http-skeleton/server/dto"
	"github.com/jonmol/http-skeleton/util/logging"
)

type Counter interface {
	IncGlobal(context.Context) (uint64, error)
	IncWord(context.Context, string) (uint64, error)
}

type Service struct {
	c Counter
}

func New(c Counter) *Service {
	return &Service{c: c}
}

func (s Service) Hello(ctx context.Context, in dto.InputHello) (*dto.OutputHello, *dto.Meta, error) {
	// gives us 2s for both calls, should be enough
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	t, err := s.c.IncGlobal(ctx)
	if err != nil {
		slog.Error("Failed to increase the global counter", logging.Err(err))
	}

	wt, err := s.c.IncWord(ctx, in.Input)
	if err != nil {
		slog.Error("Failed to increase the work counter", logging.Err(err), slog.String("word", in.Input))
	}

	if in.Input == "rude" {
		return nil, nil, dto.ErrRude
	} else if in.Input == "veryRude" {
		return nil, nil, dto.ErrVeryRude
	}

	return &dto.OutputHello{Response: fmt.Sprintf("Why hello there %s", in.Input)}, &dto.Meta{Total: t, ThisWord: wt}, nil
}
