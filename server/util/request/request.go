package request

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/schema"
	"github.com/jonmol/http-skeleton/server/util/myctx"
	"github.com/jonmol/http-skeleton/server/util/response"
	"github.com/jonmol/http-skeleton/util/logging"
)

var (
	decoder = schema.NewDecoder()
	val     = validator.New()
)

// HandlerType is the type of the business logic handler.
type HandlerFunc func(http.ResponseWriter, *http.Request) (any, any, *response.RespError)

func HandleCall(w http.ResponseWriter, r *http.Request, data any, handler HandlerFunc) {
	ctx := r.Context()
	l := myctx.LoggerFromCtx(ctx)
	if data != nil {
		switch m := r.Method; m {
		case http.MethodGet:
			if err := validateGet(ctx, data, w, r); err != nil {
				return
			}

		case http.MethodPost, http.MethodPut, http.MethodDelete:
			body, err := io.ReadAll(r.Body)
			if err != nil {
				response.JSONErrorResponse(ctx, w, response.MalformedRequest, "cannot read body")
				return
			}

			if err = json.Unmarshal(body, data); err != nil {
				l.Error("cannot unmarshal request body", logging.Err(err), slog.String("body", string(body)))
				response.JSONErrorResponse(ctx, w, response.MalformedRequest, "cannot unmarshal body")
				return
			}

			/*
					having this code instead of the four lines after will allow maps to be used as input types,
					a big problem with maps is to validate them and it'd have to happen in the handler and/or service

					if d2, ok := data.(*map[string]interface{}); ok && d2 != nil {
					   // TODO: how we want to validate maps?
					} else {
						 err = validator.New().Struct(data)
						 if err != nil {
						   response.JSONErrorResponse(ctx, w, response.MalformedRequest, err.Error())
						   return
					   }
				  }
			*/
			err = val.Struct(data)
			if err != nil {
				response.JSONErrorResponse(ctx, w, response.MalformedRequest, err.Error())
				return
			}

		default:
			response.JSONErrorResponse(ctx, w, response.Internal, "unsupported HTTP method")
			return
		}
	}

	resp, meta, httpErr := handler(w, r)
	if httpErr != nil {
		response.JSONErrorResponse(ctx, w, httpErr.Code, httpErr.Msg)
		return
	}

	response.JSONResponse(ctx, w, &response.Resp{Data: resp, Meta: meta})
}

func validateGet(ctx context.Context, data interface{}, w http.ResponseWriter, r *http.Request) error {
	l := myctx.LoggerFromCtx(ctx)
	if err := decoder.Decode(data, r.URL.Query()); err != nil {
		l.Error("cannot decode query parameters", err, "query", r.URL.Query())
		response.JSONErrorResponse(ctx, w, response.MalformedRequest, "cannot unmarshal query parameters")
		return err
	}

	if err := val.Struct(data); err != nil {
		response.JSONErrorResponse(ctx, w, response.MalformedRequest, err.Error())
		return err
	}
	return nil
}
