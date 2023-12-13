package response

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/jonmol/http-skeleton/server/util/myctx"
	"github.com/jonmol/http-skeleton/util/logging"
)

type ErrorCode string

const (
	// AlreadyExists error code
	AlreadyExists ErrorCode = "already_exists"
	// Internal error code
	Internal ErrorCode = "internal"
	// InvalidArgument error code
	InvalidArgument ErrorCode = "invalid_argument"
	// MalformedRequest error code
	MalformedRequest ErrorCode = "malformed_request"
	// MalformedResponse error code
	MalformedResponse ErrorCode = "malformed_response"
	// NotFound error code
	NotFound ErrorCode = "not_found"
	// OutOfRange error code
	OutOfRange ErrorCode = "out_of_range"
	// PermissionDenied error code
	PermissionDenied ErrorCode = "permission_denied"
	// Stale error code
	Stale ErrorCode = "stale"
	// Unauthenticated error code
	Unauthenticated ErrorCode = "unauthenticated"
	// NoContent error code
	NoContent ErrorCode = "no_content"
	// UnprocessableEntity error code
	UnprocessableEntity ErrorCode = "unprocessable_entity"
)

const (
	mimeApplicationJSON = "application/json"
	contentType         = "Content-Type"
)

var CodeMap = map[ErrorCode]int{
	AlreadyExists:       http.StatusConflict,
	Internal:            http.StatusInternalServerError,
	InvalidArgument:     http.StatusBadRequest,
	MalformedRequest:    http.StatusBadRequest,
	MalformedResponse:   http.StatusInternalServerError,
	NoContent:           http.StatusNoContent,
	NotFound:            http.StatusNotFound,
	OutOfRange:          http.StatusBadRequest,
	PermissionDenied:    http.StatusForbidden,
	Stale:               http.StatusInternalServerError,
	Unauthenticated:     http.StatusUnauthorized,
	UnprocessableEntity: http.StatusUnprocessableEntity,
}

// Resp is the response envelope. All responses from the service will allways be
// of the format
//
//	{
//	   "data": { data from the service },
//	   "meta: { meta data such as pagination, next url etc },
//	   "error": { any errors happening }
//	}
type Resp struct {
	Data  interface{} `json:"data,omitempty"`
	Meta  interface{} `json:"meta,omitempty"`
	Error *RespError  `json:"error,omitempty"`
}

// RespError is error response type
type RespError struct {
	Code ErrorCode `json:"code"`
	Msg  string    `json:"msg"`
}

// JSONResponse enforces a specific structure on all responses to make it easier to
// give a uniform response. It also enforces error codes and messages so that the
// consumer can expect a subset rather than different from each endpoint
// This would also be the place to replace the writer with a gzip writer or similar
// but I'd expect this service to be behind a load balancer or reverse proxy which will
// handle TLS certs and the Accept-Encoding header parsing
func JSONResponse(ctx context.Context, w http.ResponseWriter, res *Resp) {
	l := myctx.LoggerFromCtx(ctx)

	SetJSONContent(w)
	jr, err := json.Marshal(res)
	if err != nil {
		l.Error("JSONResponse can't unmarshal", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sc := http.StatusOK
	if res.Error != nil {
		sc = CodeMap[res.Error.Code]
	}

	w.WriteHeader(sc)

	_, err = w.Write(jr)
	if err != nil {
		l.Error("Failed to write response", logging.Err(err))
	}
}

// JSONErrorResponse implements an error json response
func JSONErrorResponse(ctx context.Context, w http.ResponseWriter, ec ErrorCode, msg string) {
	res := &Resp{
		Error: &RespError{
			Code: ec,
			Msg:  msg,
		},
	}
	JSONResponse(ctx, w, res)
}

func SetJSONContent(w http.ResponseWriter) {
	w.Header().Set(contentType, mimeApplicationJSON)
}
