package ckeys

const (
	Logger  CtxKey = "logger"
	TraceID CtxKey = "traceID"
	CtxDone CtxKey = "ctxDone"
)

type CtxKey string
