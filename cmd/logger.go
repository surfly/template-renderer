package cmd

import (
	"io"
	"log/slog"
	"os"
)

// LevelTrace is a custom slog level below Debug for maximum-detail logging.
const LevelTrace = slog.Level(-8)

// L is the application logger, configured by initLogger.
var L *slog.Logger

// initLogger configures L from the --debug/--trace flag values and returns a
// cleanup function that closes the trace file if one was opened.
func initLogger(tracePath, level string) func() {
	var w io.Writer = os.Stderr
	cleanup := func() {}
	if tracePath != "" {
		f, err := os.OpenFile(tracePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
		if err == nil {
			w = f
			cleanup = func() { _ = f.Close() }
		}
	}
	lvl := slog.LevelWarn
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "trace":
		lvl = LevelTrace
	}
	h := slog.NewTextHandler(w, &slog.HandlerOptions{Level: lvl})
	L = slog.New(h)
	slog.SetDefault(L)
	return cleanup
}
