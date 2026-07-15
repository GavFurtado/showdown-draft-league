// Package utils provides a custom slog handler and logging helpers.
//
// Logging convention (slog standard):
//
//	logger.Info("message", "key1", val1, "key2", val2)
//	logger.Warn("message", "error", err)
//	logger.Error("message", "user_id", id, "role", role)
//
// Args after the message are alternating key-value pairs.
// Always use string keys. Common keys: "error", "user_id", "role", "service", "method".
package utils

import (
	"context"
	"io"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"sync"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorDim    = "\033[2m"
)

var (
	rootLogger     *slog.Logger
	rootLoggerOnce sync.Once
)

// InitSlog initialises the root slog.Logger with a custom handler
// (positional timestamp/level, ANSI colours, key=value attrs, func=, message=).
// Must be called once at startup before any logging occurs.
func InitSlog() {
	rootLoggerOnce.Do(func() {
		handler := &customHandler{
			out:   os.Stdout,
			level: slog.LevelInfo,
		}
		rootLogger = slog.New(handler)
		slog.SetDefault(rootLogger)
	})
}

// RootLogger returns the root logger initialised by InitSlog.
func RootLogger() *slog.Logger {
	return rootLogger
}

// LoggerWithService returns a child logger with a "service" attribute attached.
//
// Usage in constructors:
//
//	logger := utils.LoggerWithService(rootLogger, "AuthService")
//
// All subsequent calls to logger.Info / logger.Warn / logger.Error / logger.Debug
// will include service=AuthService in every log line.
//
// Those methods follow slog's standard (msg string, args ...any) signature,
// where args are alternating key-value pairs:
//
//	logger.Info("user logged in", "user_id", id, "role", role)
func LoggerWithService(logger *slog.Logger, service string) *slog.Logger {
	return logger.With("service", service)
}

type customHandler struct {
	out   io.Writer
	level slog.Leveler
	attrs []slog.Attr
	mu    sync.Mutex
}

func (h *customHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

func (h *customHandler) Handle(_ context.Context, r slog.Record) error {
	buf := make([]byte, 0, 256)

	buf = r.Time.AppendFormat(buf, "2006/01/02 15:04:05")
	buf = append(buf, ' ')

	var levelColor string
	switch r.Level {
	case slog.LevelError:
		levelColor = colorRed
	case slog.LevelWarn:
		levelColor = colorYellow
	case slog.LevelInfo:
		levelColor = colorGreen
	case slog.LevelDebug:
		levelColor = colorCyan
	}
	buf = append(buf, levelColor...)
	levelStr := r.Level.String()
	buf = append(buf, levelStr...)
	for i := len(levelStr); i < 5; i++ {
		buf = append(buf, ' ')
	}
	buf = append(buf, colorReset...)
	buf = append(buf, ' ')

	for _, attr := range h.attrs {
		buf = append(buf, ' ')
		buf = h.appendAttr(buf, attr)
	}

	r.Attrs(func(a slog.Attr) bool {
		buf = append(buf, ' ')
		buf = h.appendAttr(buf, a)
		return true
	})

	if r.PC != 0 {
		frames := runtime.CallersFrames([]uintptr{r.PC})
		frame, _ := frames.Next()
		buf = append(buf, " func="...)
		trimmed := trimModulePath(frame.Function)
		buf = append(buf, trimmed...)
	}

	buf = append(buf, " message="...)
	buf = append(buf, r.Message...)

	buf = append(buf, '\n')

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.out.Write(buf)
	return err
}

func (h *customHandler) appendAttr(buf []byte, a slog.Attr) []byte {
	if a.Value.Kind() == slog.KindGroup {
		for _, attr := range a.Value.Group() {
			buf = append(buf, a.Key...)
			buf = append(buf, '.')
			buf = h.appendAttr(buf, attr)
		}
		return buf
	}
	buf = append(buf, a.Key...)
	buf = append(buf, '=')
	switch a.Key {
	case "request_id", "user_id", "league_id", "pool_entry_id",
		"draft_id", "game_id", "claim_id", "member_id",
		"discord_id", "role":
		buf = append(buf, colorDim...)
	}
	buf = append(buf, a.Value.String()...)
	switch a.Key {
	case "request_id", "user_id", "league_id", "pool_entry_id",
		"draft_id", "game_id", "claim_id", "member_id",
		"discord_id", "role":
		buf = append(buf, colorReset...)
	}
	return buf
}

func (h *customHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)
	return &customHandler{
		out:   h.out,
		level: h.level,
		attrs: newAttrs,
	}
}

func (h *customHandler) WithGroup(name string) slog.Handler {
	return h
}

const modulePrefix = "github.com/GavFurtado/showdown-draft-league/new-backend/"

func trimModulePath(fn string) string {
	return strings.TrimPrefix(fn, modulePrefix)
}
