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

func NewContextHandler(trace string, pathLogging bool) mux.MiddlewareFunc {
	traceOn := false
	if trace != "" {
		traceOn = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			l := slog.Default()
			if traceOn {
				traceID := r.Header.Get(trace)
				if traceID == "" {
					traceID = uuid.Must(uuid.NewV4()).String()
				}
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
			ctx = myctx.WithLogger(ctx, l)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
