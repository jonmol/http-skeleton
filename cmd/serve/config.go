package serve

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/jonmol/http-skeleton/cmd/config"
	"github.com/jonmol/http-skeleton/server"
)

const (
	FieldServiceName = "service-name"
	FieldDBType      = "db-type"
	FieldDBAddr      = "db-addr"
	FieldDBPass      = "db-pass"

	FieldAddress           = "http-address"
	FieldPort              = "http-port"
	FieldReadTimeout       = "http-read-timeout"
	FieldReadHeaderTimeout = "http-read-header-timeout"
	FieldIdleTimeout       = "http-idle-timeout"
	FieldMaxHeaderSize     = "http-max-header-size"
	FieldWriteTimeout      = "http-write-timeout"

	FieldTelemetry        = "telemtry"
	FieldTelemetryAddress = "telemetry-address"
	FieldTelemetryPort    = "telementry-port"

	FieldMiddlewareTraceIDHeader = "mid-trace-id-header"
	FieldMiddlewareURLPath       = "mid-url-path"

	FieldMiddlewareCors        = "mid-cors"
	FieldMiddlewareCorsOrigins = "mid-cors-origins"
	FieldMiddlewareCorsMethods = "mid-cors-methods"
	FieldMiddlewareCorsHeaders = "mid-cors-headers"

	FieldMiddlewarePromSize  = "mid-prom-size"
	FieldMiddlewarePromTime  = "mid-prom-timer"
	FieldMiddlewarePromCount = "mid-prom-counter"
)

var ConfigStructure = config.Configs{
	Ints: []config.IntConf{
		{Name: FieldPort, Desc: "Public facing http port to listen to", Def: server.DefaultPort},
		{Name: FieldTelemetryPort, Desc: "Telemetry http port to listen to", Def: server.DefaultTelemetryPort},
		{Name: FieldMaxHeaderSize, Desc: "Max header size of http requests", Def: server.DefaultMaxHeaderBytes},
	},
	Durations: []config.DurationConf{
		{Name: FieldIdleTimeout, Desc: "How long are idle keep-alive connections allowed?", Def: server.DefaultIdleTimeout},
		{Name: FieldReadTimeout, Desc: "How long to wait for data while reading HTTP requests?", Def: server.DefaultReadTimeout},
		{Name: FieldReadHeaderTimeout, Desc: "How long to wait for reading the http headers?", Def: server.DefaultReadHeaderTimeout},
		{Name: FieldWriteTimeout, Desc: "How long are HTTP writes allowed to take?", Def: server.DefaultWriteTimeout},
	},
	Strings: []config.StringConf{
		{Name: FieldServiceName, Desc: "Name of the service. Used for path and prometheus", Def: "myService"},
		{Name: FieldAddress, Desc: "Public facing address to bind to, empty for all", Def: ""},
		{Name: FieldTelemetryAddress, Desc: "Telemetry address to bind to, empty for all", Def: ""},
		{Name: FieldTelemetry, Desc: "What type of telemetry to use. prometheus|otel|none", Def: "prometheus"},
		{Name: FieldMiddlewareTraceIDHeader, Desc: "Set traceID header to be able to follow a individual request/session through the logs", Def: ""},
		{Name: FieldDBType, Desc: "What key value store to use. badger|redis", Def: "badger"},
		{Name: FieldDBAddr, Desc: "DB address", Def: filepath.Join(os.TempDir(), "http-skeleton-badger")},
		{Name: FieldDBPass, Desc: "DB password", Def: ""},
	},
	Bools: []config.BoolConf{
		{Name: FieldMiddlewareCors, Desc: "Activate CORS to allow cross domain requests from browsers", Def: true},
		{Name: FieldMiddlewareURLPath, Desc: "Add request path to the logs", Def: false},
		{Name: FieldMiddlewarePromSize, Desc: "Instrument response sizes, requires prometheus turned on to be active", Def: true},
		{Name: FieldMiddlewarePromTime, Desc: "Instrument response times, requires prometheus turned on to be active", Def: true},
		{Name: FieldMiddlewarePromCount, Desc: "Instrument request counter, requires prometheus turned on to be active", Def: true},
	},
	StringArrays: []config.StringArrayConf{
		{Name: FieldMiddlewareCorsOrigins, Desc: "List of allowed domains for CORS. See https://pkg.go.dev/github.com/jub0bs/fcors#FromOrigins for format. At least one to have CORS active.", Def: []string{"https://example.com"}},
		{Name: FieldMiddlewareCorsMethods, Desc: "List of allowed verbs for CORS requests. One or multiple of GET,HEAD,POST,PUT,PATCH,DELETE,CONNECT,OPTIONS,TRACE", Def: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}},
		{Name: FieldMiddlewareCorsHeaders, Desc: "List of allowed headers, for example Authorization", Def: []string{}},
	},
}
