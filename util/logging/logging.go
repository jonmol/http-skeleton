package logging

import (
	"log/slog"
	"runtime"
)

func Err(e error) slog.Attr {
	return slog.String("err", e.Error())
}

func Func(n string) slog.Attr {
	return slog.String("func", n)
}

func Lib(n string) slog.Attr {
	return slog.String("pkg", n)
}

// Finfo makes it easy to add file info without turning on AddSource in slog
func Finfo() slog.Attr {
	if _, filename, line, ok := runtime.Caller(1); ok {
		return slog.Group("fileInfo", slog.Int("line", line), slog.String("file", filename))
	}

	return slog.Attr{}
}
