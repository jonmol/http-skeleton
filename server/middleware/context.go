package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gofrs/uuid/v5"
	"github.com/gorilla/mux"
	"github.com/jonmol/http-skeleton/server/ckeys"
	"github.com/jonmol/http-skeleton/server/util/myctx"
)

const (
	traceIDLog = "traceID"
	pathLog    = "requestPath"
)

// NewContextHandler returns a new middleware for logging and tracing. If headerName is empty and pathLogging is false
// nothing is done apart from adding a logger to the context.
//
// If headerName is not empty the header is checked, if it has the header and it's UUID it's added to the logger so all
// logging adds the UUID for the request to the logs. This helps with tracing a single or multiple (in case the client
// supplies the header in its request) requests
//
// If pathLogging is true all log printing with the logger will add the request path to ease understanding which endpoint
// is logging.
func NewContextHandler(headerName string, pathLogging bool) mux.MiddlewareFunc {
	traceOn := false
	if headerName != "" {
		traceOn = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			if ctx.Value(ckeys.CtxDone) != nil {
				// this middleware can be added multiple times. If it is, then we want to make sure it's only adding a
				// traceID and logger to the context once, so do nothing.
				next.ServeHTTP(w, r)
			}

			l := slog.Default()
			if traceOn {
				traceID := r.Header.Get(headerName)
				if traceID != "" { // validate it, and reset if not a a uuid
					uu := uuid.FromStringOrNil(traceID)
					if uu.IsNil() {
						traceID = ""
					}
				}

				if traceID == "" {
					traceID = uuid.Must(uuid.NewV4()).String()
				}
				w.Header().Add(headerName, traceID)
				ctx = context.WithValue(r.Context(), ckeys.TraceID, traceID)
				l = slog.With(traceIDLog, traceID)
			}
			if pathLogging {
				if l != nil {
					l = l.With(pathLog, r.URL.Path)
				} else {
					l = slog.With(pathLog, r.URL.Path)
				}
			}
			ctx = context.WithValue(ctx, ckeys.CtxDone, true)
			ctx = myctx.WithLogger(ctx, l)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
