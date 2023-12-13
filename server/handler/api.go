package handler

// Handlers are only supposed to read the input, call the appropriate service function,
// check for errors and return a response. No business logic should be present
import (
	"context"
	"errors"
	"net/http"

	"github.com/jonmol/http-skeleton/server/dto"
	"github.com/jonmol/http-skeleton/server/util/myctx"
	"github.com/jonmol/http-skeleton/server/util/request"
	"github.com/jonmol/http-skeleton/server/util/response"
	"github.com/jonmol/http-skeleton/util/logging"
)

type APIService interface {
	Hello(context.Context, dto.InputHello) (*dto.OutputHello, *dto.Meta, error)
}

// Hello is a silly sample endpoint, if the parameter is "rude" it returns response.InvalidArgument, if veryRude
// it returns response.PermissionDenied otherwise 200OK and a greeting
func (h *Handler) Hello(w http.ResponseWriter, r *http.Request) {
	var body dto.InputHello
	l := myctx.LoggerFromCtx(r.Context()).With(logging.Lib("handler"))
	request.HandleCall(w, r, &body, func(w http.ResponseWriter, r *http.Request) (interface{}, interface{}, *response.RespError) {
		resp, meta, err := h.service.Hello(r.Context(), body)
		if err != nil {
			l.Error("service call failed", logging.Err(err))
			switch {
			case errors.Is(err, dto.ErrRude):
				return nil, nil, &response.RespError{Msg: err.Error(), Code: response.InvalidArgument}
			case errors.Is(err, dto.ErrVeryRude):
				return nil, nil, &response.RespError{Msg: err.Error(), Code: response.PermissionDenied}
			default:
				return nil, nil, &response.RespError{Msg: err.Error(), Code: response.Internal}
			}
		}
		if meta == nil {
			return resp, nil, nil
		}
		return resp, meta, nil
	})
}
