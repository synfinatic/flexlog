package flexlog

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
)

// Our logger which wraps slog.Logger and implements CustomLogger
type Logger struct {
	logger    *slog.Logger
	addSource bool
	color     bool
	level     *slog.LevelVar
	writer    io.Writer
}

// NewLoggerFunc creates a new Logger
func NewLogger(f NewLoggerFunc, w io.Writer, addSource bool, level slog.Leveler, color bool) *Logger {
	handle, lvl := f(w, addSource, level, color)
	return &Logger{
		logger:    slog.New(handle),
		addSource: addSource,
		color:     color,
		level:     lvl,
		writer:    w,
	}
}

// Debug logs a message at the debug level
func (l *Logger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

// DebugContext logs a message at the debug level with context
func (l *Logger) DebugContext(ctx context.Context, msg string, args ...any) {
	l.logger.DebugContext(ctx, msg, args...)
}

// Error logs a message at the error level
func (l *Logger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

// ErrorContext logs a message at the error level with context
func (l *Logger) ErrorContext(ctx context.Context, msg string, args ...any) {
	l.logger.ErrorContext(ctx, msg, args...)
}

// Info logs a message at the info level
func (l *Logger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

// InfoContext logs a message at the info level with context
func (l *Logger) InfoContext(ctx context.Context, msg string, args ...any) {
	l.logger.InfoContext(ctx, msg, args...)
}

// Log logs a message at the specified level with the included context
func (l *Logger) Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	l.logger.Log(ctx, level, msg, args...)
}

// LogAttrs logs a message at the specified level with the included context and attributes
func (l *Logger) LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	l.logger.LogAttrs(ctx, level, msg, attrs...)
}

// Warn logs a message at the warn level
func (l *Logger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

// WarnContext logs a message at the warn level with context
func (l *Logger) WarnContext(ctx context.Context, msg string, args ...any) {
	l.logger.WarnContext(ctx, msg, args...)
}

// Handler returns the slog.Handler for the logger
func (l *Logger) Handler() slog.Handler {
	return l.logger.Handler()
}

// With returns a new logger with the specified attributes
func (l *Logger) With(args ...any) *slog.Logger {
	return l.logger.With(args...)
}

// WithGroup returns a new logger with the specified group name
func (l *Logger) WithGroup(name string) *slog.Logger {
	return l.logger.WithGroup(name)
}

// Copy returns a copy of the Logger current Logger
func (l *Logger) Copy() FlexLogger {
	return NewLogger(CreateLogger, l.writer, l.addSource, l.level, l.color)
}

// Writer returns the writer for the logger
func (l *Logger) Writer() io.Writer {
	return l.writer
}

// AddSource returns whether the logger is configured to include the source file and line number in the log output
func (l *Logger) AddSource() bool {
	return l.addSource
}

// Level returns the current log level
func (l *Logger) Level() *slog.LevelVar {
	return l.level
}

// Color returns whether the logger is configured to output colorized log messages
func (l *Logger) Color() bool {
	return l.color
}

// Enabled returns whether the logger is enabled for the specified level
func (l *Logger) Enabled(ctx context.Context, level slog.Level) bool {
	return l.logger.Enabled(ctx, level)
}

// GetLogger returns the underlying slog.Logger
func (l *Logger) GetLogger() *slog.Logger {
	return l.logger
}

// SetLogger sets the underlying slog.Logger
func (l *Logger) SetLogger(logger *slog.Logger) {
	l.logger = logger
}

// SetLevel sets the log level for the logger
func (l *Logger) SetLevel(level slog.Leveler) {
	l.level.Set(level.Level())
}

// SetLevelString sets the log level for the logger by string
func (l *Logger) SetLevelString(level string) error {
	if _, ok := LevelStrings[strings.ToUpper(level)]; !ok {
		return fmt.Errorf("invalid log level: %s", level)
	}
	l.level.Set(LevelStrings[strings.ToUpper(level)].Level())
	return nil
}

// SetReportCaller sets whether to include the source file and line number in the log output
// Doing so will replace the current logger with a new one that has the new setting
func (l *Logger) SetReportCaller(reportCaller bool) {
	if l.addSource == reportCaller {
		return // do nothing
	}
	l.addSource = reportCaller
	handler, _ := CreateLogger(l.writer, l.addSource, slog.LevelWarn, l.color)
	logger.SetLogger(slog.New(handler))
}

// GetLevel returns the current log level
func (l *Logger) GetLevel() slog.Level {
	return slog.Level(l.level.Level())
}
