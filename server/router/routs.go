package router

import (
	"net/http"
)

type endpoints struct {
	health  []endpoint
	private []endpoint
	public  []endpoint
}

type endpoint struct {
	path    string
	methods []method
}

type method struct {
	verb    string
	handler func(w http.ResponseWriter, r *http.Request)
}

// ep returns all the endpoints and their config. This is where all endpoints are configured
// TODO: Add your endpoints
func ep(h IHandler) endpoints {
	return endpoints{
		health: []endpoint{
			{path: "/", methods: []method{{verb: http.MethodGet, handler: h.OK}}},
			{path: "/healthz", methods: []method{{verb: http.MethodGet, handler: h.Healthz}}},
			{path: "/readz", methods: []method{{verb: http.MethodGet, handler: h.Readyz}}},
			{path: "/livez", methods: []method{{verb: http.MethodGet, handler: h.Livez}}},
		},
		private: []endpoint{
			{path: "/hello", methods: []method{{verb: http.MethodGet, handler: h.Hello}, {verb: http.MethodOptions, handler: h.OK}}},
		},
		public: []endpoint{
			{path: "/hello", methods: []method{{verb: http.MethodGet, handler: h.Hello}, {verb: http.MethodOptions, handler: h.OK}}},
		},
	}
}
