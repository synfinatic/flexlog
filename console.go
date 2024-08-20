package flexlog

import (
	"context"
	"io"
	"log/slog"
	"runtime"
	"time"

	"github.com/lmittmann/tint"
)

const (
	FrameMarker = "__skip_frames"
)

// NewConsole creates a new slog.Handler for the ConsoleHandler, which wraps tint.NewHandler
// with some customizations.
func NewConsole(w io.Writer, addSource bool, level slog.Leveler, color bool) (slog.Handler, *slog.LevelVar) {
	lvl := new(slog.LevelVar)
	lvl.Set(level.Level())

	opts := tint.Options{
		Level:       lvl,
		AddSource:   addSource,
		ReplaceAttr: replaceAttrConsole,
		TimeFormat:  time.Kitchen,
		// TimeFormat: "",
		NoColor: true, // let the replaceAttr do the coloring
	}

	return NewConsoleHandler(w, &opts), lvl
}

// implement the slog.Handler interface via the tint.Handler
type ConsoleHandler struct {
	slog.Handler
}

// ConsoleHandler is a slog.Handler that wraps tint.Handler
func NewConsoleHandler(w io.Writer, opts *tint.Options) slog.Handler {
	return &ConsoleHandler{
		tint.NewHandler(w, opts),
	}
}

// Handle is a custom wrapper around the tint.Handler.Handle method which fixes up
// the PC value to be the correct caller for the Fatal/Trace methods
func (h *ConsoleHandler) Handle(ctx context.Context, r slog.Record) error {
	var fixStack int64 = 0
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == FrameMarker {
			fixStack = a.Value.Int64()
			return false
		}
		return true
	})

	if fixStack > 0 {
		rn := r.Clone()
		rn.PC, _, _, _ = runtime.Caller(int(fixStack))
		return h.Handler.Handle(ctx, rn)
	}
	return h.Handler.Handle(ctx, r)
}

func replaceAttrConsole(groups []string, a slog.Attr) slog.Attr {
	// Remove time from the output on the console
	if a.Key == slog.TimeKey {
		return slog.Attr{}
	}

	// Remove the frame marker attribute flag if it's present
	if a.Key == FrameMarker {
		return slog.Attr{}
	}

	// Colorize and rename the log level
	if a.Key == slog.LevelKey {
		level := a.Value.Any().(slog.Level)
		levelColor, ok := LevelColorsMap[level]
		if ok {
			a.Value = slog.StringValue(levelColor.String(logger.Color()))
		}
	}

	return a
}
