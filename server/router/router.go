package router

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jonmol/http-skeleton/server/middleware"
)

// having three layers of paths can feel a bit clunky, but if you have a load balancer and
// multiple services this will allow you to configure it to route by path. Think about
// api.example.com/v1/serviceA/private/hello and api.example.com/v1/serviceB/private/hello
// with this you can route /v1/serviceA to serviceA and /v1/serviceB to serviceB
// having version in there will also make it easier to replace parts in v2-n by routing them
// elsewhere while still offering backwards compatibility for those using /v1
const (
	versionPath = "/v1"
	privatePath = "/private"
	publicPath  = "/public"
)

var (
	serviceName string
	appPath     string
)

// Middleware is a struct of middlewares split by secure or non, where secure would
// require a JWT token or similar, while the non-secure are public facing endpoints
type Middleware struct {
	SecuredMiddleware    []mux.MiddlewareFunc
	NonSecuredMiddleware []mux.MiddlewareFunc
}

type Config struct {
	PromethusMiddlleWare bool
	PromTiming           bool
	PromCount            bool
	PromSize             bool
	Middleware           Middleware
	ServiceName          string
}

type IHandler interface {
	OK(http.ResponseWriter, *http.Request)
	Healthz(http.ResponseWriter, *http.Request)
	Readyz(http.ResponseWriter, *http.Request)
	Livez(http.ResponseWriter, *http.Request)
	Hello(http.ResponseWriter, *http.Request)
}

// BuildRouter adds all the endpoints the service should be providing
// The Prometheus middleware needs to be setup here as it needs to know all
// the paths. + is the slowest way to concatenate but it's not really a concern
// for the startup or the low amount of strings. You might want to optimize if you
// have 10k+ endpoints, but then you have other problems
func BuildRouter(han IHandler, conf Config) *mux.Router {
	serviceName = conf.ServiceName
	appPath = "/" + serviceName
	r := mux.NewRouter()

	// k8s health check endpoints
	healthCheck := r.NewRoute().Subrouter()

	eps := ep(han)

	for _, e := range eps.health {
		for _, m := range e.methods {
			healthCheck.HandleFunc(e.path, m.handler).Methods(m.verb)
		}
	}

	version := r.PathPrefix(versionPath).Subrouter()
	version.Use(conf.Middleware.NonSecuredMiddleware...)
	service := version.PathPrefix(appPath).Subrouter()

	private := service.PathPrefix(privatePath).Subrouter()
	addPromeMiddleware(conf.PromethusMiddlleWare, "private", private, eps.private, conf.PromCount, conf.PromTiming, conf.PromSize)
	private.Use(conf.Middleware.SecuredMiddleware...)
	addRoutes(private, eps.private)

	public := service.PathPrefix(publicPath).Subrouter()
	addPromeMiddleware(conf.PromethusMiddlleWare, "public", public, eps.public, conf.PromCount, conf.PromTiming, conf.PromSize)
	public.Use(conf.Middleware.NonSecuredMiddleware...)
	addRoutes(public, eps.public)

	return r
}

func addRoutes(r *mux.Router, h []endpoint) {
	for _, e := range h {
		for _, m := range e.methods {
			r.HandleFunc(e.path, m.handler).Methods(m.verb)
		}
	}
}

func addPromeMiddleware(on bool, epType string, r *mux.Router, h []endpoint, counter, timings, sizes bool) {
	if !on {
		return
	}
	baseP := versionPath + appPath + privatePath
	paths := make([]string, 0)
	for _, e := range h {
		paths = append(paths, baseP+e.path)
	}
	pMid := middleware.NewPromMiddleware(serviceName, epType, counter, sizes, timings, paths)
	r.Use(pMid)
}
