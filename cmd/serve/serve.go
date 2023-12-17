package serve

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"log/slog"

	"github.com/gorilla/mux"
	"github.com/jonmol/http-skeleton/instrumentation/otel"
	"github.com/jonmol/http-skeleton/model"
	"github.com/jonmol/http-skeleton/server"
	"github.com/jonmol/http-skeleton/server/handler"
	"github.com/jonmol/http-skeleton/server/middleware"
	"github.com/jonmol/http-skeleton/server/router"
	"github.com/jonmol/http-skeleton/server/service"
	"github.com/jonmol/http-skeleton/util/logging"
	"github.com/jub0bs/fcors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
)

type ShutdownFunc func(ctx context.Context) error

type Shutdown struct {
	name string
	fun  ShutdownFunc
}

type IServe interface {
}

type Serve struct {
	cancel       context.CancelFunc
	mut          sync.Mutex
	down         bool
	running      bool
	shutdowFuncs []Shutdown
}

// Run calls Start and waits for a shutdown signal
func (s *Serve) Run() {
	s.Start()
	stop := make(chan os.Signal, 1)

	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	sig := <-stop

	slog.Warn("Received stop signal, shutting down", "signalName", sig.String())
	s.Stop()
	s.cancel()
}

// Start starts the http server(s) and sets things up. It listens for SIGHUP and restarts if received
func (s *Serve) Start() {
	s.mut.Lock()
	defer s.mut.Unlock()

	s.shutdowFuncs = make([]Shutdown, 0, 10)

	if s.running {
		slog.Warn("Tried to start serve when already running")
		return
	}
	s.running = true
	ctx := context.Background()
	ctx, s.cancel = context.WithCancel(ctx)

	slog.Debug("Serve starting, configs", "configs", viper.AllSettings())

	if viper.GetString(FieldTelemetry) == "prometheus" {
		s.shutdowFuncs = append(s.shutdowFuncs, startInstrumentationHTTP())
	} else if viper.GetString(FieldTelemetry) == "otel" {
		shut, err := otel.SetupOTelSDK(ctx, "testing", "2.1.0")
		if err != nil {
			slog.Error("Failed to start OTEL", logging.Err(err))
			panic("Failed to setup OTEL")
		}

		s.shutdowFuncs = append(s.shutdowFuncs, Shutdown{name: "otel", fun: shut})
	}
	db := connectDB(ctx)
	if err := db.EnsureDB(ctx); err != nil {
		slog.Error("Failed to setup the db!", logging.Err(err))
		panic(err)
	}

	s.shutdowFuncs = append(s.shutdowFuncs, Shutdown{"db", db.Close}, startAPIHTTP(db))

	s.sigHUP()
}

// sigHUP restarts the server, which will re-read all configs
func (s *Serve) sigHUP() {
	go func() {
		stop := make(chan os.Signal, 1)

		signal.Notify(stop, syscall.SIGHUP)
		sig := <-stop

		slog.Warn("Received SIGHUP, restarting", "signalName", sig.String())
		s.Stop()
		s.Start()
	}()
}

// startInstrumentationHTTP starts a separate http.Server on port FieldTelemetryPort. The reason for a separate one is to make
// it less likely to accidentally expose the /metrics path
func startInstrumentationHTTP() Shutdown {
	ser := server.New(viper.GetDuration(FieldReadTimeout),
		viper.GetDuration(FieldReadHeaderTimeout),
		viper.GetDuration(FieldWriteTimeout),
		viper.GetDuration(FieldIdleTimeout),
		viper.GetInt(FieldTelemetryPort),
		viper.GetInt(FieldMaxHeaderSize),
		viper.GetString(FieldTelemetryAddress),
	)

	go func() {
		if err := ser.Start(promhttp.Handler()); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				slog.Info("Instrumentation HTTP Server stopped")
			} else {
				slog.Error("Failed to start the http server", err)
				panic("Without HTTP it makes little sense to continue")
			}
		}
	}()
	return Shutdown{"instrumentation-http-server", ser.Stop}
}

// startAPIHTTP configures and starts the API http server. It adds middlewares that will be used for the endpoints
func startAPIHTTP(db *model.DB) Shutdown {
	// setup an HTTP listener
	ser := server.New(viper.GetDuration(FieldReadTimeout),
		viper.GetDuration(FieldReadHeaderTimeout),
		viper.GetDuration(FieldWriteTimeout),
		viper.GetDuration(FieldIdleTimeout),
		viper.GetInt(FieldPort),
		viper.GetInt(FieldMaxHeaderSize),
		viper.GetString(FieldAddress),
	)

	// setup routes and start serving HTTP
	go func() {
		route := setupRouter(db)
		if err := ser.Start(route); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				slog.Info("HTTP Server stopped")
			} else {
				slog.Error("Failed to start the http server", err)
				panic("Without HTTP it makes little sense to continue")
			}
		}
	}()
	return Shutdown{"api-http-server", ser.Stop}
}

func connectDB(ctx context.Context) *model.DB {
	db := model.NewModel(ctx)
	switch viper.GetString(FieldDBType) {
	case "badger":
		if err := db.OpenBadger(ctx, viper.GetString(FieldDBAddr)); err != nil {
			panic(fmt.Sprintf("Failed to open %s at %s", viper.GetString(FieldDBType), viper.GetString(FieldDBAddr)))
		} else {
			slog.Info("Connected to Badger", slog.String("path", viper.GetString(FieldDBAddr)))
		}
	case "redis":
		if err := db.OpenRedis(ctx, viper.GetString(FieldDBAddr), viper.GetString(FieldDBPass)); err != nil {
			panic(fmt.Sprintf("Failed to connect to %s at %s", viper.GetString(FieldDBType), viper.GetString(FieldDBAddr)))
		} else {
			slog.Info("Connected to Redis", slog.String("path", viper.GetString(FieldDBAddr)))
		}
	default:
		panic(fmt.Sprintf("Unsupported database %s selected.", viper.GetString(FieldDBType)))
	}
	return db
}

func setupRouter(db *model.DB) *mux.Router {
	serviceS := service.New(db.Counter)
	han := handler.New(serviceS)

	mid := router.Middleware{SecuredMiddleware: addSecMiddlewares(), NonSecuredMiddleware: addPublicMiddlewares()}

	rConf := router.Config{
		Middleware:  mid,
		PromCount:   viper.GetBool(FieldMiddlewarePromCount),
		PromSize:    viper.GetBool(FieldMiddlewarePromSize),
		PromTiming:  viper.GetBool(FieldMiddlewarePromTime),
		ServiceName: viper.GetString(FieldServiceName),
	}

	if viper.GetString(FieldTelemetry) == "prometheus" {
		rConf.PromethusMiddlleWare = true
	}
	return router.BuildRouter(han, rConf)
}

// addSecMiddlewares adds any middlewares to be used on secure endpoints
func addSecMiddlewares() []mux.MiddlewareFunc {
	mid := make([]mux.MiddlewareFunc, 0, 2)
	mid = append(mid, middleware.NewContextHandler(viper.GetString(FieldMiddlewareTraceIDHeader), viper.GetBool(FieldMiddlewareURLPath)))

	if viper.GetBool(FieldMiddlewareCors) &&
		len(viper.GetStringSlice(FieldMiddlewareCorsOrigins)) > 0 &&
		len(viper.GetStringSlice(FieldMiddlewareCorsMethods)) > 0 {
		o := viper.GetStringSlice(FieldMiddlewareCorsOrigins)
		m := viper.GetStringSlice(FieldMiddlewareCorsMethods)
		h := viper.GetStringSlice(FieldMiddlewareCorsHeaders)

		var (
			cors func(http.Handler) http.Handler
			err  error
		)

		if len(h) > 0 { // since headers isn't mandatory and fcors require "one, rest..." as arguments we can't just call it like the others
			cors, err = fcors.AllowAccess(
				fcors.FromOrigins(o[0], o[1:]...),
				fcors.WithMethods(m[0], m[1:]...),
				fcors.WithRequestHeaders(h[0], h[1:]...),
			)
		} else {
			cors, err = fcors.AllowAccess(
				fcors.FromOrigins(o[0], o[1:]...),
				fcors.WithMethods(m[0], m[1:]...),
			)
		}
		if err != nil {
			slog.Error("Failed to load cors", logging.Err(err))
			panic("Failed to initialize the cors middleware. All is in vain, giving up")
		}

		mid = append(mid, cors)
	} else {
		slog.Warn("CORS check disabled, do you really want it like that?")
	}
	return mid
}

func addPublicMiddlewares() []mux.MiddlewareFunc {
	mid := []mux.MiddlewareFunc{middleware.NewContextHandler(viper.GetString(FieldMiddlewareTraceIDHeader), viper.GetBool(FieldMiddlewareURLPath))}

	return mid
}

// Stop gracefully shuts down anything started.
func (s *Serve) Stop() {
	slog.Info("Starting shutdown")
	s.mut.Lock()
	defer s.mut.Unlock()

	// server is already shut down
	if s.down {
		slog.Info("Server already shutting down")
		return
	}
	s.down = true

	slog.Info("Calling all shutdown functions")

	for _, service := range s.shutdowFuncs {
		slog.Info(fmt.Sprintf("Shutting down %s", service.name))
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		if err := service.fun(ctx); err != nil {
			slog.Error("Failed to shutdown %s", slog.String("serviceName", service.name), logging.Err(err))
		}

		cancel()
	}

	// calling the global ctx cancel function to cancel anything dangling
	s.cancel()
	s.running = false

	slog.Info("Shutdown complete")
}
