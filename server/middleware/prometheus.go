package middleware

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/jonmol/http-skeleton/util/logging"
	"github.com/prometheus/client_golang/prometheus"
)

// TrackedWriter is just adding a bytes counter to http.ResponseWriter
// to be able to
type TrackedWriter struct {
	http.ResponseWriter
	Bytes int
}

func (t *TrackedWriter) Write(b []byte) (int, error) {
	n, err := t.ResponseWriter.Write(b)
	t.Bytes += n
	return n, err
}

func (t *TrackedWriter) WriteHeader(code int) {
	t.ResponseWriter.WriteHeader(code)
	var buf bytes.Buffer
	err := t.Header().Write(&buf)
	if err != nil {
		slog.Error("Failed to write response code", logging.Err(err))
	}
	t.Bytes += buf.Len()
}

// NewPromMiddleware creates a new Prometheus middleware. It excludes routes with {pathVariables} as they
// become "infinitely" many with for instance userID in the path. It can meassure 3 values:
// request count per endpoint (counter)
// response size per endpoint (gauge)
// response time per endpoint (gauge)
// one or two can be disabled, but it will panic if all three are, then you shouldn't use it
func NewPromMiddleware(appname, endpointType string, counted, sized, timed bool, paths []string) mux.MiddlewareFunc {
	if !counted && !sized && !timed {
		panic("All three measures are off, turn off the middleware instead!")
	}

	// only instrument paths without path variables
	validP := make([]string, 0, len(paths))
	for _, p := range paths {
		if !strings.Contains(p, "{") {
			validP = append(validP, p)
		}
	}

	counter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: fmt.Sprintf("%s_%s_endpoint_counter", appname, endpointType),
		Help: "Counter for the endpoints",
	},
		[]string{"endpoint"},
	)

	sizes := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: fmt.Sprintf("%s_%s_endpoint_sizes", appname, endpointType),
		Help: "The size of the responses",
	}, []string{"endpoint"})

	times := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: fmt.Sprintf("%s_%s_endpoint_times", appname, endpointType),
		Help: "The time it takes to send a response",
	}, []string{"endpoint"})

	registerMetrics(times, sizes, counter, timed, sized, counted, validP)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			inc(counter, counted, r.URL.Path)
			d := time.Now()
			tw := &TrackedWriter{ResponseWriter: w}
			if sized {
				next.ServeHTTP(tw, r)
			} else {
				next.ServeHTTP(w, r)
			}
			observe(times, timed, r.URL.Path, time.Since(d))
			observe(sizes, sized, r.URL.Path, tw.Bytes)
		})
	}
}

func inc(counter *prometheus.CounterVec, active bool, path string) {
	if active {
		if c, err := counter.GetMetricWith(prometheus.Labels{"endpoint": path}); err == nil {
			c.Inc()
		}
	}
}

func observe[V int | time.Duration](h *prometheus.HistogramVec, active bool, path string, val V) {
	if active {
		if c, err := h.GetMetricWith(prometheus.Labels{"endpoint": path}); err == nil {
			c.Observe(float64(val))
		}
	}
}

func registerMetrics(t, s *prometheus.HistogramVec, c *prometheus.CounterVec, ta, sa, ca bool, paths []string) {
	if ca {
		prometheus.DefaultRegisterer.MustRegister(c)
	}
	if sa {
		prometheus.DefaultRegisterer.MustRegister(s)
	}
	if ta {
		prometheus.DefaultRegisterer.MustRegister(t)
	}

	for _, p := range paths {
		if ca {
			c.With(prometheus.Labels{"endpoint": p})
		}
		if sa {
			s.With(prometheus.Labels{"endpoint": p})
		}
		if ta {
			t.With(prometheus.Labels{"endpoint": p})
		}
	}
}
