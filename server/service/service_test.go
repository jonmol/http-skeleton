package service_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/jonmol/http-skeleton/server/dto"
	"github.com/jonmol/http-skeleton/server/service"
	mocks "github.com/jonmol/http-skeleton/server/service/mocks"
	"github.com/jonmol/http-skeleton/server/util/myctx"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type testFunc func(context.Context)

func withDefaults(_ *testing.T, f testFunc) {
	ctx := myctx.WithLogger(context.Background(), slog.Default())
	f(ctx)
}

func TestUnitHello(t *testing.T) {
	r := require.New(t)

	withDefaults(t, func(ctx context.Context) {
		tests := []struct {
			name    string           // name of the test to be able to follow it
			input   dto.InputHello   // input data
			resData *dto.OutputHello // service data back
			resErr  error            // service error back
			dbResG  uint64
			dbErrG  error
			dbResW  uint64
			dbErrW  error
		}{
			{name: "no input", input: dto.InputHello{}, resData: &dto.OutputHello{Response: "Why hello there "}, resErr: nil},                                                                                       // no input
			{name: "proper input", input: dto.InputHello{Input: "world"}, resData: &dto.OutputHello{Response: "Why hello there world"}, resErr: nil, dbResG: 1, dbResW: 2},                                          // proper input
			{name: "proper input failed db1", input: dto.InputHello{Input: "world"}, resData: &dto.OutputHello{Response: "Why hello there world"}, resErr: nil, dbErrG: errors.New("e1"), dbErrW: errors.New("e2")}, // proper input both db fail
			{name: "proper input failed db2", input: dto.InputHello{Input: "world"}, resData: &dto.OutputHello{Response: "Why hello there world"}, resErr: nil, dbErrG: errors.New("some db err")},                  // proper input global fail
			{name: "proper input failed db3", input: dto.InputHello{Input: "world"}, resData: &dto.OutputHello{Response: "Why hello there world"}, resErr: nil, dbErrW: errors.New("some db err")},                  // proper input word fail
			{name: "rude", input: dto.InputHello{Input: "rude"}, resData: nil, resErr: dto.ErrRude},                                                                                                                 // rude
			{name: "very rude", input: dto.InputHello{Input: "veryRude"}, resData: nil, resErr: dto.ErrVeryRude},                                                                                                    // very rude
		}
		for _, test := range tests {
			test := test // since we run the tests in parallel
			t.Run(test.name, func(t *testing.T) {
				t.Parallel()
				count := mocks.NewMockCounter(t)
				count.EXPECT().IncGlobal(mock.Anything).Return(test.dbResG, test.dbErrG)
				count.EXPECT().IncWord(mock.Anything, mock.Anything).Return(test.dbResW, test.dbErrW)

				s := service.New(count)

				res, _, err := s.Hello(context.Background(), test.input)
				t.Logf("Running test case %s", test.name) // log the name, only visible when failing so it's possible to know which test fails
				r.Equal(test.resData, res)
				r.Equal(test.resErr, err)
			})
		}
	})
}
