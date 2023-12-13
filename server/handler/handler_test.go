package handler_test

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jonmol/http-skeleton/server/dto"
	"github.com/jonmol/http-skeleton/server/handler"
	mocks "github.com/jonmol/http-skeleton/server/handler/mocks"

	"github.com/jonmol/http-skeleton/server/util/myctx"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type testFunc func(context.Context)

var helloResponses = []string{
	`{"error":{"code":"internal","msg":"unexpected error"}}`,
	`{"data":{"response":"Why hello there world"}}`,
	`{"error":{"code":"invalid_argument","msg":"no response to rude people"}}`,
	`{"error":{"code":"permission_denied","msg":"outrageous input"}}`,
	`{"error":{"code":"malformed_request","msg":"Key: 'InputHello.Input' Error:Field validation for 'Input' failed on the 'required' tag"}}`,
}

func withDefaults(_ *testing.T, f testFunc) {
	ctx := myctx.WithLogger(context.Background(), slog.Default())
	f(ctx)
}

func TestUnitHello(t *testing.T) {
	r := require.New(t)

	withDefaults(t, func(ctx context.Context) {
		tests := []struct {
			name     string           // name of the test to be able to follow it
			req      string           // request string
			svcData  *dto.OutputHello // service data back
			svcMeta  *dto.Meta        // service meta data back
			svcErr   error            // service error back
			respCode int              // expected response code
			respData string           // expected response data
			noCall   bool             // in the case of no input, the service isn't called
		}{
			{name: "weird error", req: "world", svcData: nil, svcErr: errors.New("unexpected error"), respCode: http.StatusInternalServerError, respData: helloResponses[0]},                    // unexpected service response
			{name: "proper input", req: "world", svcData: &dto.OutputHello{Response: "Why hello there world"}, svcMeta: nil, svcErr: nil, respCode: http.StatusOK, respData: helloResponses[1]}, // proper input
			{name: "rude", req: "rude", svcData: nil, svcErr: dto.ErrRude, respCode: http.StatusBadRequest, respData: helloResponses[2]},                                                        // rude input
			{name: "very rude", req: "veryRude", svcData: nil, svcErr: dto.ErrVeryRude, respCode: http.StatusForbidden, respData: helloResponses[3]},                                            // very rude input
			{name: "no input", req: "", svcData: &dto.OutputHello{}, svcErr: nil, respCode: http.StatusBadRequest, respData: helloResponses[4], noCall: true},                                   // no input
		}

		for _, test := range tests {
			test := test
			t.Run(test.name, func(t *testing.T) {
				t.Parallel()
				// mock the service, if test.noCall is true, it means we don't expect it to call the service (in this case because input validation fails so the service isn't called)
				s := mocks.NewMockService(t)
				if test.noCall {
					s.AssertNotCalled(t, "Hello")
				} else {
					s.EXPECT().Hello(mock.Anything, mock.Anything).Return(test.svcData, test.svcMeta, test.svcErr)
				}

				// create a test request with test.req as the parameter
				req := httptest.NewRequest(http.MethodGet, "http://localhost:8000/v1/xray_chat/send_message?input="+test.req, http.NoBody)
				req = req.WithContext(ctx)
				respW := httptest.NewRecorder()

				// create a handler with the mocked service and call Hello with the test req and writer
				h := handler.New(s)
				h.Hello(respW, req)

				r.Equal(test.respCode, respW.Code)
				r.Equal(test.respData, respW.Body.String())
			})
		}
	})
}
